# histdata.com import

Import prices (ticks) from histdata.com for backtesting. They offer data for forex, gold, and SP500.

[Full list of supported instruments](http://www.histdata.com/download-free-forex-data/?/ascii/tick-data-quotes).

## Usage

### Download CSV files

Change period and instrument in histdata.rb:

```ruby
for i_date in 2020..2021 # change date
```

and 

```ruby
fxpair = 'SPXUSD' # change your instrument
```

Then run the script. CSV files should be downloading now.

```shell
./cmd/import-histdata/histdata.rb
```

### Unzip files

```shell
mv HISTDATA* data/
find ./data/ -name 'HISTDATA*zip' -exec unzip {} \;
```

### Import 
Now run the importer which generates 1min candles and stores them to local sqlite DB:

```shell
INSTRUMENT="SPXUSD" IMPORT_HISTDATA_CSV_FILES=`ls *.csv | tr "\n" ","` go run cmd/import-histdata/main.go
```

Then you can run the backtesting tool to use the data with the `.db` file in `./data`.

## TODOs

- Remove ruby script and support downloading CSV files in the Go program.