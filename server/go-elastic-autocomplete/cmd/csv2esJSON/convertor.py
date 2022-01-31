# Installation command : pip3 install elasticsearch
# Run command: python3 convertor.py
from elasticsearch import Elasticsearch, helpers
import csv

# Create the elasticsearch client.
es = Elasticsearch(host = "localhost", port = 9200)

# Open csv file and bulk upload
with open('cmd/csv2esJSON/finalList.csv') as f:
    reader = csv.DictReader(f)
    helpers.bulk(es, reader, index='test')

