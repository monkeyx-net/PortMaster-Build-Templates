#!/bin/bash
clear
green='\033[32m'
nocolour='\033[0m'

modify_port_sh () {
port_name="$1"
folder_name="$2"
port_script="$3"


}


create_new_port () {
  port=$1
  read -rp  "Your portname is $port. Do you wish to change the title. Y/N "  answer
  case "$answer" in
    [yY]|[yY][eE][sS])
        read -rp "Enter PortName: " port
        ;;
    [nN]|[nN][oO])
        echo "Exiting..."
        ;;
    *)
        echo "Invalid input, exiting..."
        exit 1
        ;;
  esac
  echo $port
  read -rp "zip/folder name: " folder_name
  folder_name=$(echo "$folder_name" | tr '[:upper:]' '[:lower:]')
  read -rp "Enter Bash file name: " bash_file
  cp new_ports/template/zz_Port.sh new_ports/${folder_name}/${bash_file}
}

exit_nicely () {
  tput sgr0
}

# move back to root folder of repo
cd ..

if ! command -v jq &> /dev/null; then
    echo "Error: jq is not installed. Please install jq to continue."
    echo "On Arch: sudo pacman -S jq"
    echo "On Ubuntu/Debian: sudo apt-get install jq"
    echo "On CentOS/RHEL: sudo yum install jq"
    exit 1
fi


echo -e "${green}"
echo -e "*******************************************"
echo -e "* Welcome to the PortMaster recipe system *"
echo -e "*******************************************"


echo -e "\n\nThis script should be run from the recipes folder to ensure correct paths are used. \n\n"

read -rp 'Enter port name or part name: ' PORT




# Case insensitive search "i"
mapfile -t ports < <(
  jq -r --arg search "$PORT" '
    .ports[]
    | select(.attr.title? != null and (.attr.title | test($search; "il")))
    | .attr.title
  ' releases/ports.json
)

if [ ${#ports[@]} -eq 0 ]; then
    echo "No ports found with ${PORT} in title"
    echo "Please add basic port information - See the wiki https://github.com/monkeyx-net/PortMaster-Build-Templates/wiki"
    echo "Process single port file instead. Need path some folder/shamogu"
    echo "Make function for port.json and ports.json"
    echo "Make function for returning title and name ie shamogu.zip"
    create_new_port ${PORT}
fi

echo "Available ports with ${PORT} in the title:"
for i in "${!ports[@]}"; do
    echo "$((i + 1))) ${ports[i]}"
done

while true; do
    echo -n "Enter choice: "
    read choice
    if [[ $choice =~ ^[0-9]+$ ]] && [ $choice -ge 1 ] && [ $choice -le ${#ports[@]} ]; then
        port_title="${ports[$((choice-1))]}"
        break
    else
        echo "Invalid selection! Please enter a number between 1 and ${#ports[@]}"
    fi
done
echo "Now processing: $port_title"

jq '.ports[] | select(.attr.title == "Manic*") | .name' releases/ports.json

echo "Enter URL for port code base. Must be in .zip or .tar.gz format"
read -rp 'Enter URL for code ' input
if [ -z "$input" ]; then
    echo "Empty input received"
elif [[ "$input" =~ \.zip$ || "$input" =~ \.tar\.gz$ ]]; then
    echo "Valid URL ending with .zip or .tar.gz"
else
    echo "URL must end with .zip or .tar.gz"
fi

exit_nicely

