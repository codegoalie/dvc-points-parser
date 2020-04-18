#/bin/sh

set -e

for i in ~/Documents/DVC/2021/point-charts/*.pdf; do
  ~/Downloads/pdftotext -table $i converted-charts/2021/$(basename -- $i).txt
done
for i in ~/Documents/DVC/2020/point-charts/*.pdf; do
  ~/Downloads/pdftotext -table $i converted-charts/2020/$(basename -- $i).txt
done
