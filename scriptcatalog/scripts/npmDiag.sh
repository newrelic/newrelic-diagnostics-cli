#!/usr/bin/env bash

collectPreReqCheck() {

	# Checks for root user
	if [ "$EUID" -ne 0 ]; then
	    echo "Please run this script as root, or with sudo to use this function."
	    exit 0
	fi

	# Check to make sure `jq` is installed
	if ! [[ $(command -v jq) ]]; then
	    echo "Please install the \`jq\` package to continue."
	    exit 0
	fi

	# Check to make sure `zip` is installed
	if ! [[ $(command -v zip) ]]; then
	    echo "Please install the \`zip\` package to continue."
	    exit 0
	fi
}

walkPreReqCheck() {
	# Check to make sure `yq` is installed
	if ! [[ $(command -v yq) ]]; then
	    echo "Please install the \`yq\` package to continue."
	    exit 0
	fi

	# Check to make sure `snmpwalk` is installed
	if ! [[ $(command -v snmpwalk) ]]; then
	    echo "Please install \`snmpwalk\` to continue."
	    echo "This can be done with \`apt-get install snmp\` or \`yum install net-snmp-utils\`."
	    exit 0
	fi

	# Check to make sure `jq` is installed
	if ! [[ $(command -v jq) ]]; then
	    echo "Please install the \`jq\` package to continue."
	    exit 0
	fi
}

collectRoutine() {

	# Finds container IDs using Ktranslate image. Adds them to `foundContainerIDs` array.
	readarray -t allContainerIDs < <(docker ps -aq)
	for i in "${allContainerIDs[@]}"; do
		if [[ $(docker inspect "$i" | jq -r '.[] | .Config | .Image') == 'kentik/ktranslate:v2' ]]; then
			foundContainerIDs+=("$i")
		fi
	done

	# Displays menu and allows container selection.
	echo "=-=-=-=-=-=-=-=-=-=-="
	for i in "${!foundContainerIDs[@]}"; do
		echo "[$i] $(docker inspect "${foundContainerIDs[i]}" | jq -r '.[] | .Name' | sed 's/^\///')"
	done
	echo ""
	echo "[q] Exit npmDiag"
	echo "=-=-=-=-=-=-=-=-=-=-="

	# Forces selection to be space-delimited list of integers, or a q to exit
	# Outputs `menuSelectedOption` array
	menuLoopExit=false
	while [[ "$menuLoopExit" == false ]]; do
		echo ""
		echo "Enter space-delimited list of containers you want diagnostic data from (0 1 2...)"
		read -ep "You can also use 'q' to exit the script > " menuSelection
		if [[ "$menuSelection" =~ ^[0-9]+($|[[:space:]]|[0-9]+){1,}$ ]]; then
			read -a menuSelectedOption <<< "$menuSelection"
			for i in "${menuSelectedOption[@]}"; do
				if [[ "$i" -ge "${#foundContainerIDs[@]}" ]]; then
					echo ""
					echo "Container selection must include only integers shown in the listed options."
					menuLoopExit=false
					unset menuSelectedOption
					break
				else
					menuLoopExit=true
				fi
			done
		elif [[ "${menuSelection,,}" = "q" ]]; then
			echo ""
			echo "Exiting..."
			exit 0
		else
			echo ""
			echo "Selection must be a space-delimited list of integers, or a 'q' to exit"
		fi
	done

	# Parses menu selection. Builds `targetContainerIDs` array.
	for i in "${menuSelectedOption[@]}"; do
		targetContainerIDs+=("${foundContainerIDs[i]}")
	done

	# Creates working directory in `/tmp`. 
	tmpFolderName="npmDiag-$(date +%s)"
	mkdir /tmp/"$tmpFolderName"
	for i in "${targetContainerIDs[@]}"; do

		# Gets config file for each container. Copies to a `/tmp` working folder. Renames with Docker container long ID.
		if [[ -f $(docker inspect "$i" | jq -r '.[] | .Mounts | .[] | .Source | select(contains("yaml"))') ]]; then
			cp $(docker inspect "$i" | jq -r '.[] | .Mounts | .[] | .Source | select(contains("yaml"))') /tmp/"$tmpFolderName"/$(docker inspect "$i" | jq -r '.[] | .Id').yaml
		fi

		# Stops container, deletes old log file, starts container to regenerate logs. Copies to a `/tmp` working folder. Skips if logs are missing to begin with.		
		if [[ -f $(docker inspect "$i" | jq -r '.[] | .LogPath') ]]; then
			echo ""
			echo "Stopping $(docker inspect "$i" | jq -r '.[] | .Name' | sed 's/^\///')"
			docker stop "$i" > /dev/null
			rm $(docker inspect "$i" | jq -r '.[] | .LogPath')
			echo "Regenerating log file..."
			docker start "$i" > /dev/null

			# Sleeps script for 3 minutes to allow container logs to regenerate
			timer=180
			while [[ "$timer" -gt 0 ]]; do
			    minutes=$((timer / 60))
			    seconds=$((timer % 60))
			    printf "\rTime remaining: %02d:%02d" $minutes $seconds
			    sleep 1
			    timer=$((timer - 1))
			done
			printf "\rTime remaining: %02d:%02d\n" 0 
			echo "Done recreating logs for $(docker inspect "$i" | jq -r '.[] | .Name' | sed 's/^\///')"
			cp $(docker inspect "$i" | jq -r '.[] | .LogPath') /tmp/"$tmpFolderName"/$(docker inspect "$i" | jq -r '.[] | .Id').log
		else
			echo ""
			echo "Skipping logs for $(docker inspect "$i" | jq -r '.[] | .Name' | sed 's/^\///'). Original log file missing."
		fi

		# Gets `inspect` for each container. Copies to a `/tmp` working folder.
		docker inspect "$i" > /tmp/"$tmpFolderName"/$(docker inspect "$i" | jq -r '.[] | .Id')-$(docker inspect "$i" | jq -r '.[] | .Name' | sed 's/^\///').dockerInspect.out
	done
}

diagZip() {
	# Bundles output files into zip file and places it in the $HOME directory
    zip -qj npmDiag-output.zip /tmp/"$tmpFolderName"/*
    chmod 666 npmDiag-output.zip
    echo ""
    echo "Created output file \`npmDiag-output.zip\` in working directory."
}

postCleanup() {
	# Deletes files from `/tmp`
    rm -rd /tmp/"$tmpFolderName" > /dev/null 2>&1
}

walkRoutine() {

	# Finds container IDs using Ktranslate image. Adds them to `foundContainerIDs` array.
	readarray -t allContainerIDs < <(docker ps -aq)
	for i in "${allContainerIDs[@]}"; do
		if [[ $(docker inspect "$i" | jq -r '.[] | .Config | .Image') == 'kentik/ktranslate:v2' ]]; then
			foundContainerIDs+=("$i")
		fi
	done

	# Displays menu and allows container selection.
	echo "=-=-=-=-=-=-=-=-=-=-="
	for i in "${!foundContainerIDs[@]}"; do
		echo "[$i] $(docker inspect "${foundContainerIDs[i]}" | jq -r '.[] | .Name' | sed 's/^\///')"
	done
	echo ""
	echo "[q] Exit npmDiag"
	echo "=-=-=-=-=-=-=-=-=-=-="

	# Forces selection to be a single integer, or a q to exit
	# Outputs `targetContainerID`
	menuLoopExit=false
	while [[ "$menuLoopExit" == false ]]; do
		echo ""
		echo "Enter integer value for the container running SNMP collection on target device"
		read -ep "You can also use 'q' to exit the script > " menuSelection
		if [[ "$menuSelection" =~ ^[0-9]{1,}$ ]]; then

			if [[ "$menuSelection" -ge "${#foundContainerIDs[@]}" ]]; then
				echo ""
				echo "Selection must be a single integer from the container list, or a 'q' to exit"
				unset menuSelection
			else
				targetContainerID="${foundContainerIDs[$menuSelection]}"
				menuLoopExit=true
			fi
		elif [[ "${menuSelection,,}" = "q" ]]; then
			echo ""
			echo "Exiting..."
			exit 0
		else
			echo ""
			echo "Selection must be a single integer from the container list, or a 'q' to exit"
		fi
	done

	# Sets location of target container's config file as `configLocation`
	configLocation=$(docker inspect "$targetContainerID" | jq -r '.[] | .Mounts | .[] | .Source | select(contains("yaml"))')

	# Displays menu and allows device selection
	readarray -t availableDevices < <(yq eval '.devices.*.device_name' "$configLocation")
	echo ""
	echo "=-=-=-=-=-=-=-=-=-=-="
	for i in "${!availableDevices[@]}"; do
		echo "[$i] ${availableDevices[$i]}"
	done
	echo ""
	echo "[q] Exit npmDiag"
	echo "=-=-=-=-=-=-=-=-=-=-="

	# Forces selection to be a single integer, or a q to exit
	# Outputs `targetDeviceName`
	menuLoopExit=false
	unset menuSelection
	while [[ "$menuLoopExit" == false ]]; do
		echo ""
		echo "Enter integer value the target device"
		read -ep "You can also use 'q' to exit the script > " menuSelection
		if [[ "$menuSelection" =~ ^[0-9]{1,}$ ]]; then

			if [[ "$menuSelection" -ge "${#availableDevices[@]}" ]]; then
				echo ""
				echo "Selection must be a single integer from the device list, or a 'q' to exit"
				unset menuSelection
			else
				targetDeviceName="${availableDevices[$menuSelection]}"
				menuLoopExit=true
			fi
		elif [[ "${menuSelection,,}" = "q" ]]; then
			echo ""
			echo "Exiting..."
			exit 0
		else
			echo ""
			echo "Selection must be a single integer from the device list, or a 'q' to exit"
		fi
	done

	# Checks for v3 credentials inside the device's config. Runs v2 if they're not found
	if [[ $(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .snmp_v3' "$configLocation") == "null" ]]; then
		# Runs snmpwalk routine for v2 devices
		# Uses `yq` to parse file for values at runtime instead of storing as variables
		# This is done to avoid issues with escaping special characters. Not the fastest route but it works.
		echo "Running full snmpwalk on '$targetDeviceName'. This could take upward of 10 minutes."
		targetDeviceIP=$(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .device_ip' "$configLocation")
		snmpwalk -v 2c -On -c $(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .snmp_comm' "$configLocation") "$targetDeviceIP" . >> "$targetDeviceName"-snmpwalk.out
		echo ""
		echo "Created output file \`"$targetDeviceName"-snmpwalk.out\` in working directory."
	else
		# Runs snmpwalk routine for v3 devices
		targetDeviceIP=$(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .device_ip' "$configLocation")
		targetDeviceAuthProtocol=$(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .snmp_v3 | .authentication_protocol' "$configLocation")
		targetDevicePrivProtocol=$(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .snmp_v3 | .privacy_protocol' "$configLocation")

		# Uses `yq` to parse file for values at runtime instead of storing as variables
		# This is done to avoid issues with escaping special characters. Not the fastest route but it works.
		if [[ "$targetDeviceAuthProtocol" == "NoAuth" && "$targetDevicePrivProtocol" == "NoPriv" ]]; then
			echo "Running full NoAuthNoPriv snmpwalk on '$targetDeviceName'. This could take upward of 10 minutes."
			snmpwalk -v3 -On -l noAuthNoPriv -u $(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .snmp_v3 | .user_name' "$configLocation") "$targetDeviceIP" . >> "$targetDeviceName"-snmpwalk.out

		elif [[ "$targetDeviceAuthProtocol" != "NoAuth" && "$targetDevicePrivProtocol" == "NoPriv" ]]; then
			echo "Running full AuthNoPriv snmpwalk on '$targetDeviceName'. This could take upward of 10 minutes."
			snmpwalk -v3 -On -l authNoPriv -u $(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .snmp_v3 | .user_name' "$configLocation") -a "$targetDeviceAuthProtocol" -A $(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .snmp_v3 | .authentication_passphrase' "$configLocation") "$targetDeviceIP" . >> "$targetDeviceName"-snmpwalk.out

		elif [[ "$targetDeviceAuthProtocol" != "NoAuth" && "$targetDevicePrivProtocol" != "NoPriv" ]]; then
			echo "Running full AuthPriv snmpwalk on '$targetDeviceName'. This could take upward of 10 minutes."
			snmpwalk -v3 -On -l authPriv -u $(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .snmp_v3 | .user_name' "$configLocation") -a "$targetDeviceAuthProtocol" -A $(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .snmp_v3 | .authentication_passphrase' "$configLocation") -x "$targetDevicePrivProtocol" -X $(yq eval '.devices | .[] | select(.device_name == "'"$targetDeviceName"'") | .snmp_v3 | .privacy_passphrase' "$configLocation") "$targetDeviceIP" . >> "$targetDeviceName"-snmpwalk.out

		fi

	fi
}

# Checks number of arguments
if [ $# -ne 1 ]; then
    echo "Usage: $0 [--collect|--walk|--help]"
    exit 1
fi

# Validates argument passed to script
if [ "$1" != "--collect" ] && [ "$1" != "--walk" ] && [ "$1" != "--help" ]; then
    echo "Invalid argument: $1"
    echo "Usage: $0 [--collect|--walk|--help]"
    exit 1
fi

# Executes a routine based on passed argument
if [ "$1" = "--collect" ]; then
    collectPreReqCheck
    collectRoutine
    diagZip
    postCleanup
elif [ "$1" = "--walk" ]; then
    walkPreReqCheck
    walkRoutine
else
    echo "Usage: npmDiag [--collect|--walk|--help]"
    echo "       --collect: Collects diagnostic info from containers. Outputs a zip file called \`npmDiag-output.zip\`"
    echo "       --walk: Run \`snmpwalk\` against a device from the config. Outputs \`<deviceName>-snmpwalk.out\`"
    echo "       --help: Shows this help message."
fi
