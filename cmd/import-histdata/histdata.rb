#!/usr/bin/env ruby

# This script only support ASCII format
# and tick download.

require 'uri'
require 'net/http'
require 'mechanize'

agent = Mechanize.new
agent.pluggable_parser.default = Mechanize::Download

for i_date in 2020..2021 # change date
  date = i_date.to_s
  for i_month in 1..12 # change month
    month = '%02d' % i_month
    datemonth = date + month

    fxpair = 'SPXUSD' # change your instrument
    platform = 'ASCII'
    timeframe = 'T'
    saved_name = 'HISTDATA_COM_' + [platform, fxpair, timeframe, datemonth].join('_') + '.zip'
    referer_uri = "http://www.histdata.com/download-free-forex-historical-data/?/#{platform.downcase}/tick-data-quotes/#{fxpair.downcase}/#{date}/#{month}"

    puts "Downloading: #{saved_name}"

    if File.exist?saved_name
      puts "Downloaded: #{saved_name}, skipping."
      next
    end

    p = agent.get(referer_uri)
    next_p = p.form_with(:name => 'file_down').submit
    next_p.save(saved_name)
  end
end