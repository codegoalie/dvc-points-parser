package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/codegoalie/dvc-points-parser/api"
	"github.com/codegoalie/dvc-points-parser/models"
	"github.com/codegoalie/dvc-points-parser/resorts"
	"github.com/urfave/cli"
)

const dryRun = false

var parsedResorts []resorts.Resort

// const serverURL = "http://localhost:3000/v1"
const serverURL = "https://api.lineleader.io/v1"

// Result mapsa trip finding result
type Result struct {
	resort         resorts.Resort
	roomType       resorts.RoomType
	pointBlock     resorts.PointBlock
	possibleNights int
	totalPoints    int
}

func (r Result) String() string {
	return fmt.Sprintf(
		"%-57s  %-5s %-5s  %-30s  %-20s  %4d  %4d\n",
		r.resort.Name,
		r.pointBlock.CheckInAt.Format("1/2"),
		r.pointBlock.CheckOutAt.Format("1/2"),
		r.roomType.Name,
		r.roomType.ViewType,
		r.possibleNights,
		r.totalPoints,
	)
}

// Results are many Result
type Results []Result

func (r Results) Len() int {
	return len(r)
}

func (r Results) Less(i, j int) bool {
	return r[i].pointBlock.CheckInAt.Before(r[j].pointBlock.CheckInAt)
}

func (r Results) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func main() {
	app := &cli.App{
		Name:  "points-parser",
		Usage: "Parse DVC points charts and upload to API",
		// Flags: flags,
		// Action: func(c *cli.Context) error {
		// 	facts.PrintAFact(webhookURL)
		// 	return nil
		// },
		Commands: []cli.Command{
			{
				Name: "parse-one",
				Action: func(c *cli.Context) error {
					args := c.Args()
					if len(args) < 1 {
						return fmt.Errorf("Please specify a file to parse")
					}

					fmt.Println("Parsing one", args)

					f, err := os.Open(args.First())
					if err != nil {
						return fmt.Errorf("failed to open file: %w", err)
					}

					resort, err := resorts.ParseUneditedFile(f, "2025")
					if err != nil {
						err = fmt.Errorf("failed to parse file: %w", err)
						log.Println(err)
						return err
					}

					for _, roomType := range resort.RoomTypes {
						fmt.Println(roomType.Name, roomType.ViewType)
						// fmt.Printf("%v\n", points)
					}

					return nil
				},
			},
			{
				Name: "parse-all",
				Action: func(c *cli.Context) error {
					var err error
					client := api.New(serverURL)
					remotes, err := client.GetResorts()
					if err != nil {
						err = fmt.Errorf("failed to get resorts from API: %w", err)
						log.Fatal(err)
					}

					targetPoints := 180
					minNights := 5
					results := Results{}

					resortMapping := map[string]map[string]map[string]int32{}
					err = json.Unmarshal([]byte(resortMapJSON), &resortMapping)
					if err != nil {
						err = fmt.Errorf("failed to unmarshal resort map: %w", err)
						log.Fatal(err)
					}
					fmt.Println(resortMapping)
					insertStatement := strings.Builder{}
					insertStatement.WriteString(
						"INSERT INTO travel_periods (room_type_id, start_on, end_on, weekday_points, weekend_points) VALUES \n",
					)

					// root := "vgf/"
					root := "converted-charts/2025/"
					// // root := "converted-charts/2022/FINAL_2022_DVC_VGF_Pt_Chts.pdf.txt"
					err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
						if info.IsDir() {
							return nil
						}

						fmt.Println(path)
						resort, err := resorts.ParseFile(path)
						fmt.Println("r", resort)
						if err != nil {
							err = fmt.Errorf("failed to parse files: %w", err)
							log.Println(err)
							return err
						}

						var ok bool
						idMap, ok := resortMapping[resort.Name]
						if !ok {
							log.Fatal("Resort not found in map:", resort.Name)
						}

						// log.Printf("%+v\t%d\n", resort.Name, len(resort.RoomTypes))

						remote := remoteByName(resort, remotes)
						if remote == nil {
							log.Println("Failed to find resort by name", resort.Name)
							return nil
						}
						// log.Println(remote)

						remoteRooms, err := client.GetRoomTypes(remote.ID)
						if err != nil {
							err = fmt.Errorf("failed to get room types for %s: %w", remote.Name, err)
							return err
						}

						// if strings.Contains(resort.Name, "Vero") ||
						// 	strings.Contains(resort.Name, "Aulani") ||
						// 	strings.Contains(resort.Name, "Saratoga") ||
						// 	strings.Contains(resort.Name, "Hilton") {
						// 	return nil
						// }

						for _, roomType := range resort.RoomTypes {
							typeMap, ok := idMap[roomType.Name]
							if !ok {
								log.Fatal("Failed to find room type in map", roomType.Name)
							}
							var id int32
							if id, ok = typeMap[roomType.ViewType]; !ok {
								log.Fatalf("Failed to find view in map '%s' %s'", roomType.Name, roomType.ViewType)
							}
							fmt.Println(resort.Name, roomType.Name, roomType.ViewType, id)

							// if strings.Contains(roomType.Name, "Studio") {
							// 	continue
							// }

							// if strings.Contains(resort.Name, "Animal") &&
							// 	(strings.Contains(roomType.ViewType, "Standard") ||
							// 		strings.Contains(roomType.ViewType, "Value")) {
							// 	continue
							// }
							remoteRoom := remoteRoom(roomType, remoteRooms)
							if remoteRoom == nil {
								log.Printf("%+v\n", remoteRooms)
								log.Println("Failed to find remote room type", remote.Name, roomType.Name, roomType.ViewType)
								continue
							}

							for _, pointBlock := range roomType.PointChart {
								insertStatement.WriteString(
									fmt.Sprintf(
										"(%d, '%s', '%s', %d, %d),\n",
										id,
										pointBlock.CheckInAt.Format("2006-01-02"),
										pointBlock.CheckOutAt.Format("2006-01-02"),
										pointBlock.WeekdayPoints,
										pointBlock.WeekendPoints,
									),
								)
								daysInBlock := int(pointBlock.CheckOutAt.Sub(pointBlock.CheckInAt).Hours() / 24)
								totalPoints := 0
								var nightsPossible int
								for nightsPossible = 0; nightsPossible < 31 && nightsPossible <= daysInBlock; nightsPossible++ {
									dow := time.Weekday(nightsPossible % 7)
									var newPoints int
									if dow == time.Friday || dow == time.Saturday {
										newPoints = pointBlock.WeekendPoints
									} else {
										newPoints = pointBlock.WeekdayPoints
									}

									if totalPoints+newPoints > targetPoints {
										break
									}
									totalPoints += newPoints
								}

								if nightsPossible >= minNights {
									results = append(results, Result{
										resort:         resort,
										roomType:       roomType,
										pointBlock:     pointBlock,
										possibleNights: nightsPossible,
										totalPoints:    totalPoints,
									})
								}
								reqs := []api.PointsReq{}
								currentStayOn := pointBlock.CheckInAt
								for {

									thesePoints := pointBlock.WeekdayPoints
									if currentStayOn.Weekday() == time.Friday || currentStayOn.Weekday() == time.Saturday {
										thesePoints = pointBlock.WeekendPoints
									}

									reqs = append(reqs, api.PointsReq{
										RoomTypeID: remoteRoom.ID,
										StayOn:     currentStayOn.ToTime(),
										Amount:     thesePoints,
									})
									// fmt.Println(req)

									currentStayOn = currentStayOn.AddDate(0, 0, 1)

									if currentStayOn.After(pointBlock.CheckOutAt) {
										break
									}
								}
								if !dryRun {
									client.CreatePoints(reqs)
								}
							}
						}

						return nil
					})
					if err != nil {
						panic(err)
					}

					sort.Sort(results)
					fmt.Println(insertStatement.String())

					// for _, result := range results {
					// 	fmt.Printf("%s", result)
					// }
					// fmt.Println(len(results))
					return nil
				},
			},
		},
	}
	app.UseShortOptionHandling = true

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func remoteByName(local resorts.Resort, remotes []models.Resort) *models.Resort {
	for _, remote := range remotes {
		if local.Name == remote.Name {
			return &remote
		}
	}

	return nil
}

func remoteRoom(local resorts.RoomType, remotes []models.Room) *models.Room {
	for _, remote := range remotes {
		if local.Name == remote.Name && local.ViewType == remote.ViewType {
			return &remote
		}
	}

	return nil
}

const resortMapJSON = `{
	"Aulani, Disney Vacation Club Villas, Ko Olina, Hawaii": {
		"Deluxe Studio": {
			"Island Gardens View": 3,
			"Ocean View": 5,
			"Poolside Gardens View": 4,
			"Standard View": 2
		},
		"Hotel Room": {
			"Standard View": 1
		},
		"One-Bedroom Villa": {
			"Island Gardens View": 7,
			"Ocean View": 9,
			"Poolside Gardens View": 8,
			"Standard View": 6
		},
		"Three-Bedroom Grand Villa": {
			"Ocean View": 19,
			"Standard View": 18
		},
		"Two-Bedroom Lock-Off Villa": {
			"Island Gardens View": 12,
			"Ocean View": 16,
			"Poolside Gardens View": 14,
			"Standard View": 10
		},
		"Two-Bedroom Villa": {
			"Island Gardens View": 13,
			"Ocean View": 17,
			"Poolside Gardens View": 15,
			"Standard View": 11
		}
	},
	"Bay Lake Tower at Disney's Contemporary Resort": {
		"Deluxe Studio": {
			"Lake View": 66,
			"Standard View": 65,
			"Theme Park View": 67
		},
		"One-Bedroom Villa": {
			"Lake View": 69,
			"Standard View": 68,
			"Theme Park View": 70
		},
		"Three-Bedroom Grand Villa": {
			"Lake View": 77,
			"Theme Park View": 78
		},
		"Two-Bedroom Lock-Off Villa": {
			"Lake View": 75,
			"Standard View": 74,
			"Theme Park View": 76
		},
		"Two-Bedroom Villa": {
			"Lake View": 72,
			"Standard View": 71,
			"Theme Park View": 73
		}
	},
	"Boulder Ridge Villas at Disney's Wilderness Lodge": {
		"Deluxe Studio": {
			"Standard View": 116
		},
		"One-Bedroom Villa": {
			"Standard View": 117
		},
		"Two-Bedroom Lock-Off Villa": {
			"Standard View": 119
		},
		"Two-Bedroom Villa": {
			"Standard View": 118
		}
	},
	"Copper Creek Villas & Cabins at Disney's Wilderness Lodge": {
		"Cabin": {
			"Standard View": 37
		},
		"Deluxe Studio": {
			"Standard View": 30,
			"with Walk-In Shower": 31
		},
		"One-Bedroom Villa": {
			"Standard View": 32
		},
		"Three-Bedroom Grand Villa": {
			"Standard View": 36
		},
		"Two-Bedroom Lock-Off Villa": {
			"Standard View": 34,
			"with Walk-In Shower": 35
		},
		"Two-Bedroom Villa": {
			"Standard View": 33
		}
	},
	"Disney's Animal Kingdom Villas - Jambo House": {
		"Deluxe Studio": {
			"Kilimanjaro Club Concierge": 41,
			"Savanna View": 40,
			"Standard View": 39,
			"Value Accommodation": 38
		},
		"One-Bedroom Villa": {
			"Kilimanjaro Club Concierge": 45,
			"Savanna View": 44,
			"Standard View": 43,
			"Value Accommodation": 42
		},
		"Three-Bedroom Grand Villa": {
			"Savanna View": 50
		},
		"Two-Bedroom Lock-Off Villa": {
			"Kilimanjaro Club Concierge": 49,
			"Savanna View": 48,
			"Standard View": 47,
			"Value Accommodation": 46
		}
	},
	"Disney's Animal Kingdom Villas - Kidani Village": {
		"Deluxe Studio": {
			"Savanna View": 52,
			"Standard View": 51
		},
		"One-Bedroom Villa": {
			"Savanna View": 54,
			"Standard View": 53
		},
		"Three-Bedroom Grand Villa": {
			"Savanna View": 60,
			"Standard View": 59
		},
		"Two-Bedroom Lock-Off Villa": {
			"Savanna View": 58,
			"Standard View": 57
		},
		"Two-Bedroom Villa": {
			"Savanna View": 56,
			"Standard View": 55
		}
	},
	"Disney's Beach Club Villas": {
		"Deluxe Studio": {
			"Standard View": 61
		},
		"One-Bedroom Villa": {
			"Standard View": 62
		},
		"Two-Bedroom Lock-Off Villa": {
			"Standard View": 64
		},
		"Two-Bedroom Villa": {
			"Standard View": 63
		}
	},
	"Disney's BoardWalk Villas": {
		"Deluxe Studio": {
			"Boardwalk View": 21,
			"Garden/Pool View": 22,
			"Standard View": 20
		},
		"One-Bedroom Villa": {
			"Boardwalk View": 24,
			"Garden/Pool View": 25,
			"Standard View": 23
		},
		"Three-Bedroom Grand Villa": {
			"Boardwalk View": 29
		},
		"Two-Bedroom Lock-Off Villa": {
			"Boardwalk View": 27,
			"Garden/Pool View": 28,
			"Standard View": 26
		}
	},
	"Disney's Hilton Head Island Resort": {
		"Deluxe Studio": {
			"Standard View": 95
		},
		"One-Bedroom Villa": {
			"Standard View": 96
		},
		"Three-Bedroom Grand Villa": {
			"Standard View": 98
		},
		"Two-Bedroom Villa": {
			"Standard View": 97
		}
	},
	"Disney's Old Key West Resort": {
		"Deluxe Studio": {
			"Near Hospitality House": 100,
			"Standard View": 99
		},
		"One-Bedroom Villa": {
			"Near Hospitality House": 102,
			"Standard View": 101
		},
		"Three-Bedroom Grand Villa": {
			"Near Hospitality House": 108,
			"Standard View": 107
		},
		"Two-Bedroom Lock-Off Villa": {
			"Near Hospitality House": 106,
			"Standard View": 105
		},
		"Two-Bedroom Villa": {
			"Near Hospitality House": 104,
			"Standard View": 103
		}
	},
	"Disney's Polynesian Villas & Bungalows": {
		"Deluxe Studio": {
			"Lake View": 121,
			"Standard View": 120
		},
		"Two-Bedroom Bungalow": {
			"Lake View": 122
		}
	},
	"Disney's Riviera Resort": {
		"Deluxe Studio": {
			"Preferred View": 125,
			"Standard View": 124
		},
		"One-Bedroom Villa": {
			"Preferred View": 127,
			"Standard View": 126
		},
		"Three-Bedroom Grand Villa": {
			"Preferred View": 132
		},
		"Tower Studio": {
			"Standard View": 123
		},
		"Two-Bedroom Lock-Off Villa": {
			"Preferred View": 131,
			"Standard View": 130
		},
		"Two-Bedroom Villa": {
			"Preferred View": 129,
			"Standard View": 128
		}
	},
	"Disney's Saratoga Springs Resort & Spa": {
		"Deluxe Studio": {
			"Preferred": 80,
			"Standard": 79
		},
		"One-Bedroom Villa": {
			"Preferred": 82,
			"Standard": 81
		},
		"Three-Bedroom Grand Villa": {
			"Preferred": 88,
			"Standard": 87
		},
		"Treehouse Villa": {
			"Standard": 89
		},
		"Two-Bedroom Lock-Off Villa": {
			"Preferred": 86,
			"Standard": 85
		},
		"Two-Bedroom Villa": {
			"Preferred": 84,
			"Standard": 83
		}
	},
	"Disney's Vero Beach Resort": {
		"Deluxe Inn Room Ocean View": {
			"Standard View": 111
		},
		"Deluxe Inn Room Standard View": {
			"Standard View": 109
		},
		"Deluxe Studio": {
			"Standard View": 110
		},
		"One-Bedroom Villa": {
			"Standard View": 112
		},
		"Three-Bedroom Beach Cottage": {
			"Standard View": 115
		},
		"Two-Bedroom Lock-Off Villa": {
			"Standard View": 114
		},
		"Two-Bedroom Villa": {
			"Standard View": 113
		}
	},
	"The Villas at Disney's Grand Californian Hotel & Spa": {
		"Deluxe Studio": {
			"Standard View": 90
		},
		"One-Bedroom Villa": {
			"Standard View": 91
		},
		"Three-Bedroom Grand Villa": {
			"Standard View": 94
		},
		"Two-Bedroom Lock-Off Villa": {
			"Standard View": 93
		},
		"Two-Bedroom Villa": {
			"Standard View": 92
		}
	},
	"The Villas at Disney's Grand Floridian Resort & Spa": {
		"Deluxe Studio": {
			"Lake View": 134,
			"Standard View": 133
		},
		"One-Bedroom Villa": {
			"Lake View": 136,
			"Standard View": 135
		},
		"Resort Studio": {
			"Lake View": 143,
			"Standard View": 142,
			"Theme Park": 144
		},
		"Three-Bedroom Grand Villa": {
			"Lake View": 141
		},
		"Two-Bedroom Lock-Off Villa": {
			"Lake View": 140,
			"Standard View": 138
		},
		"Two-Bedroom Villa": {
			"Lake View": 139,
			"Standard View": 137
		}
	},
	"The Villas at Disneyland Hotel": {
		"Duo Studio": {
			"Standard View": 145,
			"Preferred View": 146,
			"Garden Room": 147
		},
		"Deluxe Studio": {
			"Standard View": 148,
			"Preferred View": 149,
			"Garden Room": 150
		},
		"One-Bedroom Villa": {
			"Preferred View": 151
		},
		"Two-Bedroom Lock-Off Villa": {
			"Preferred View": 152
		},
		"Two-Bedroom Villa": {
			"Preferred View": 153
		},
		"Three-Bedroom Grand Villa": {
			"Preferred View": 154
		}
	},
	"The Cabins at Disney's Fort Wilderness Resort - A Disney Vacation Club Resort": {
		"Cabin": {
			"Standard View": 155
		}
	}
}`
