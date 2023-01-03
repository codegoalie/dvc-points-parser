#/bin/sh

set -e

for i in ~/Documents/DVC/2024/point-charts/*.pdf; do
  ~/Downloads/pdftotext -table $i converted-charts/2024/$(basename -- $i).txt
done
# for i in ~/Documents/DVC/2023/point-charts/*.pdf; do
#   ~/Downloads/pdftotext -table $i converted-charts/2023/$(basename -- $i).txt
# done
# for i in ~/Documents/DVC/2022/point-charts/*.pdf; do
#   ~/Downloads/pdftotext -table $i converted-charts/2022/$(basename -- $i).txt
# done
# for i in ~/Documents/DVC/2021/point-charts/*.pdf; do
#   ~/Downloads/pdftotext -table $i converted-charts/2021/$(basename -- $i).txt
# done
# for i in ~/Documents/DVC/2020/point-charts/*.pdf; do
#   ~/Downloads/pdftotext -table $i converted-charts/2020/$(basename -- $i).txt
# done
