#!/usr/bin/env bash

# init arrays
declare -a TASKS    # holds list of tasks
declare -a EXPLAINS # holds list of explains

# read each task file and populate the arrays. Set IFS to avoid breaking on spaces.
# if debugging this script, to see the list of tasks and explains collected, run these 2 commands:
#  find ./tasks -mindepth 2 -type f -name '*.go' -exec awk -F'"' '/Explain\(\)/ { getline; print FILENAME ": " $2 }' {} + > explains.out
#  find ./tasks -mindepth 2 -type f -name '*.go' -exec grep IdentifierFromString {} + > tasks.out
IFS=$'\n'
while read -r file; do
	if [[ -n "${file##*example*}" ]] && [[ -n "${file##*_test.go}" ]]; then # filter out example files and tests
		TASKS+=($(grep IdentifierFromString ${file} | cut -d\" -f 2))
		EXPLAINS+=($(awk -F'"' '/Explain\(\)/ { getline; print $2 }' ${file})) # easier to do with awk than grep since we need the next line
	fi
done <<<"$(find ./tasks -mindepth 2 -type f -name '*.go')"
unset IFS # unset IFS

# count num tasks and explains gathered, it should match, exit with error if they don't
NUMTASKS=${#TASKS[@]}
NUMEXPLAINS=${#EXPLAINS[@]}

if [[ $NUMTASKS != $NUMEXPLAINS ]]; then
	echo "Number of tasks (${NUMTASKS}) does not equal number of explains (${NUMEXPLAINS})"

	exit 1 # exit with error
fi

echo -e "\nTotal number of tasks: ${NUMTASKS}"

# some variables for the loop below
PREVIOUSCATEGORY="" # keep track of which category we are on
COUNT=1             # keep count for each category

# loops through the tasks and explains and build the markdown
for ((i = 0; i < ${NUMTASKS}; i++)); do
	CATEGORY=(${TASKS[$i]%%/*}) # %%/* returns first string split on / aka the category

	if [[ ${CATEGORY,,} != ${PREVIOUSCATEGORY,,} ]]; then
		echo -e "\n### ${CATEGORY}\n" # add the category header
		COUNT=1                       # reset count
	fi

	echo "${COUNT}. **${TASKS[$i]}** - ${EXPLAINS[$i]}" # output the task and explain
	PREVIOUSCATEGORY=(${CATEGORY})                      # keep track of which category we are on
	COUNT=$((COUNT + 1))
done

# exit gracefully
exit 0
