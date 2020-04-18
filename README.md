# DVC Points Parser

Converts [`pdftotext`](https://www.xpdfreader.com/pdftotext-man.html) converted
DVC points charts into a [`badger`](https://github.com/dgraph-io/badger) 
database for performing calculations upon; such as
[`dvc-points-calculator`](https://github.com/codegoalie/dvc-points-calculator).

Already converted 2020-2021 points charts are available in `converted-charts/`.
`process.sh` converts years worth of points charts to text format.
