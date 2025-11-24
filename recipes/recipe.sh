#!/bin/bash
# move back to root folder of repo
cd ..
clear
green='\033[32m'
nocolour='\033[0m'


update_port_json() {
    local port_title="$1"
    local porter_name="$2"
    local port_folder="$3"
    local port_bash="$4"

    jq --arg port_title "$port_title" \
       --arg porter_name "$porter_name" \
       --arg port_folder "$port_folder" \
       --arg port_bash "$port_bash" \
       '.name = $port_folder + ".zip" |
        .items[0] = $port_bash |
        .items[1] = $port_folder |
        .attr.title = $port_title |
        .attr.porter[0] = $porter_name' new_ports/${port_folder}/${port_folder}/port.json > new_ports/${port_folder}/${port_folder}/port.tmp && mv new_ports/${port_folder}/${port_folder}/port.tmp new_ports/${port_folder}/${port_folder}/port.json
}


create_new_port () {
  local port_title=$1
  echo -e "\n\n"
  read -rp  "Your portname is ${port_title}. Do you wish to change the title. Y/N "  answer
  case "$answer" in
    [yY]|[yY][eE][sS])
        read -rp "Enter PortName: " port_title
        ;;
    [nN]|[nN][oO])
        ;;
    *)
        echo "Invalid input, exiting..."
        exit 1
        ;;
  esac
  echo -e "\nCreating a recipe file for port title: ${port_title}"
  echo -e "This will create basic files for a new port"
  read -rp "Enter PortMaster name for your port: " porter_name
  read -rp "Enter port exectuable name: " port_exe
  read -rp "Enter zip/folder name: " port_folder
  folder_name=$(echo "$port_folder" | tr '[:upper:]' '[:lower:]')
  read -rp "Enter Bash file name (include .sh): " port_bash
  if [[ "${port_bash: -3}" != ".sh" ]]; then
    port_bash="${port_bash}.sh"
    echo "Added .sh extension: $port_bash"
  fi
  mkdir -p new_ports/${port_folder}
  cp new_ports/templates/port_template/zz_Port.sh new_ports/${port_folder}/${port_bash}
  sed -i 's/zz_port/'"${port_exe}"'/g; s/zz_folder/'"${port_folder}"'/g' new_ports/${port_folder}/${port_bash}
  cp -r new_ports/templates/port_template/zz_folder/ new_ports/${port_folder}/${port_folder}
  mv new_ports/${port_folder}/${port_folder}/zz_folder.gptk new_ports/${port_folder}/${port_folder}/${port_exe}.gtpk
  mv new_ports/${port_folder}/${port_folder}/zz_folder_LICENSE.txt new_ports/${port_folder}/${port_folder}/${port_exe}_LICENSE.txt
  mv new_ports/${port_folder}/${port_folder}/zz_folder.md new_ports/${port_folder}/${port_folder}/${port_exe}.md
  update_port_json $port_title $porter_name $port_folder $port_bash
}

exit_nicely () {
  tput sgr0
}


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

read -rp 'Enter port name or part name: ' port_title



# Case insensitive search "i"
mapfile -t ports < <(
  jq -r --arg search "$port_title" '
    .ports[]
    | select(.attr.title? != null and (.attr.title | test($search; "il")))
    | .attr.title
  ' releases/ports.json
)

if [ ${#ports[@]} -eq 0 ]; then
    echo -e "\n\nNo ports found with ${port_title} in title"
    echo "Please add basic port information - See the wiki https://github.com/monkeyx-net/PortMaster-Build-Templates/wiki"
    echo "Process single port file instead. Need path some folder/shamogu"
    echo "Make function for port.json and ports.json"
    echo "Make function for returning title and name ie shamogu.zip"
    create_new_port ${port_title}
fi

echo "Available ports with ${port_title} in the title:"
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

