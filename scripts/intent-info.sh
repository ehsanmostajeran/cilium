#!/bin/bash

curl -XGET "http://localhost:9200/_search?q=$1&pretty"
