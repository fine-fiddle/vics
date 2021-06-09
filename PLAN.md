# Plan

## Goal V1

CLI tool.

- this will be the technology that powers a different app- maybe ios, maybe web, maybe native. Maybe alfred or whathaveyou.
Only really handles emoji- small file size

## "Architecture"

Sqlite stores the images as base64, along with metadata.
python to do the programming, because we know it. Python is a compromise choice, because something with a binary will be much nicer.

## Database Architecture

One Table- "all".
SHA of Image | Image B64 | name | filetype? or does this go in b64 | name tags | source tags | date

## "Operation" "Interface"

### Ingester

- receive image
- convert to b64
- ask for metadata tags
- store in database
- analyze image dimensions?

### Search

- Filter based on tags

### Yeild file somehow
