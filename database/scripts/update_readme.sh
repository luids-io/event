#!/bin/bash

SRCDIR=$(dirname $(readlink -f "$0"))
DBDIR="$(dirname "$SRCDIR")"
OUTPUT="$DBDIR/README.adoc"
URLBASE="."

cat > $OUTPUT <<EOF
= Event Database

This is a database of event codes ready to be used on an
link:https://github.com/luids-io/event[event server].

EOF

function genTableEvents() {
	echo "== Registered events" >> $OUTPUT
	echo "" >> $OUTPUT
	
	echo "|===" >> $OUTPUT
	echo "| Code | Type | Codename | Description | Tags | Fields" >> $OUTPUT
	
	jq -s '[.[][]]' ${DBDIR}/*.json | jq -s '.[][]' \
		| jq -r '{ code: .code, type: .type, codename: .codename, desc: .description, tags: (.tags // empty) |join(" "), fields: .fields | map(.name? // empty) | join(" ")}' \
		| jq -r '[.code, .type, .codename, .desc, .tags, .fields] | @tsv' |
	while IFS=$'\t' read -r code stype codename desc tags fields; do
		sourcefile=`grep "$code" ${DBDIR}/*.json | cut -f1 -d ":"`
		sourcefile=`basename "$sourcefile"`
	
		echo "" >> $OUTPUT
		echo "|link:$URLBASE/${sourcefile}[$code]" >> $OUTPUT
		echo "|$stype" >> $OUTPUT
		echo "|$codename" >> $OUTPUT
		echo "|$desc" >> $OUTPUT
		echo "|$tags" >> $OUTPUT
		echo "|$fields" >> $OUTPUT
	done
	echo "|===" >> $OUTPUT
}

echo "[[events-table]]" >> $OUTPUT
genTableEvents

