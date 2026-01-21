#!/bin/bash

GITHUB_URL="https://github.com/monkeyx-net/PortMaster-Build-Templates/actions/workflows/buildrecipe.yml"

exit_nicely () {
  tput sgr0
  exit
}

read_check() {
    local prompt="${1}"
    local var_name="${2}"

    while true; do
        read -rp "${prompt}" input
        if [ -n "${input}" ]; then
            eval "${var_name}='${input}'"
            break
        else
            echo "Input cannot be empty. Please try again."
        fi
    done
}

create_recipe (){
local port_folder="${1}"
local port_url="${2}"
local port_build="${3}"
local port_checksum="${4}"
local port_exe="${5}"
local port_pre="${port_folder}_pre.sh"
local port_build="${port_folder}_build.sh"
local port_post="${port_folder}_post.sh"
mkdir -p "recipes/files/${1}"
echo '#!/bin/bash' > "recipes/files/${1}/${port_pre}"
echo '# Use this script for pre build configuration steps' >> "recipes/files/${1}/${port_pre}"
echo '#!/bin/bash' > "recipes/files/${1}/${port_build}"
echo '# Use this script for build commands.' >> "recipes/files/${1}/${port_build}"
echo '#!/bin/bash' > "recipes/files/${1}/${port_post}"
echo '# Use this script for post build commands and getting binaries and resources ready to create artifacts' >> "recipes/files/${1}/${port_post}"


jq -n \
  --arg port_name "${port_folder}.zip" \
  --arg port_folder "${port_folder}" \
  --arg port_updated "${port_date}" \
  --arg port_url "${port_url}" \
  --arg port_build "${port_build}" \
  --arg port_checksum "${port_checksum}" \
  --arg port_exe "${port_exe}" \
  --arg port_pre "${port_pre}" \
  --arg port_build "${port_build}" \
  --arg port_post "${port_post}" \
  '{
    name: $port_name,
    port_json: ("new_ports/" + $port_folder + "/" + $port_folder + "/port.json"),
    source: {
      date_updated: $port_updated,
      sha256: $port_checksum,
      port_exe: $port_exe,
      port_build: $port_build,
      port_url: $port_url,
      port_pre: $port_pre,
      port_build: $port_build,
      port_post: $port_post
    }
  }' > recipes/files/${port_folder}_recipe.json

}
update_port_json() {
    local port_title="${1}"
    local porter_name="${2}"
    local port_folder="${3}"
    local port_bash="${4}"
    local port_pre="${port_folder}_pre.sh"
    local port_build="${port_folder}_build.sh"
    local port_post="${port_folder}_post.sh"

    jq --arg port_title "$port_title" \
       --arg porter_name "$porter_name" \
       --arg port_folder "$port_folder" \
       --arg port_bash "${port_bash}" \
       '.name = $port_folder + ".zip" |
        .items[0] = $port_bash |
        .items[1] = $port_folder |
        .attr.title = $port_title |
        .attr.porter[0] = $porter_name' new_ports/${port_folder}/${port_folder}/port.json > new_ports/${port_folder}/${port_folder}/port.tmp
        mv new_ports/${port_folder}/${port_folder}/port.tmp new_ports/${port_folder}/${port_folder}/port.json
}

create_new_port () {
  local port_title="${1}"
  echo -e "\n\n"
  read -rp  "Your portname is ${port_title}. Do you wish to change the title. Y/N "  answer
  case "${answer}" in
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
  read_check "Enter PortMaster porter name ie github, discord or nickname for your port: " porter_name
  read_check "Enter port executable name: " port_exe
  read_check "Enter zip/folder name: " port_folder
  folder_name=$(echo "${port_folder}" | tr '[:upper:]' '[:lower:]')
  read_check "Enter Bash file name (include .sh): " port_bash
  if [[ "${port_bash:(-3)}" != ".sh" ]]; then
    port_bash="${port_bash}.sh"
    echo "Added .sh extension: ${port_bash}"
  fi
  mkdir -p new_ports/${port_folder}
  cp recipes/templates/port_template/zz_Port.sh "new_ports/${port_folder}/${port_bash}"
  sed -i 's/zz_port/'"${port_exe}"'/g; s/zz_folder/'"${port_folder}"'/g' "new_ports/${port_folder}/${port_bash}"
  cp -r recipes/templates/port_template/zz_folder/ new_ports/${port_folder}/${port_folder}
  mv new_ports/${port_folder}/${port_folder}/zz_folder.gptk new_ports/${port_folder}/${port_folder}/${port_exe}.gptk
  mv new_ports/${port_folder}/${port_folder}/zz_folder_LICENSE.txt new_ports/${port_folder}/${port_folder}/${port_exe}_LICENSE.txt
  update_port_json "${port_title}" "${porter_name}" "${port_folder}" "${port_bash}"
  while true; do
    echo "Enter URL for port code base. Must be in .zip or .tar.gz format"
    read_check 'Enter URL for code: ' port_url
    if [ -z "${port_url}" ]; then
      echo "Empty input received. Please enter a valid URL."
    elif [[ "${port_url}" =~ \.zip$ || "${port_url}" =~ \.tar\.gz$ ]]; then
      echo "Creating sha256sum for ${port_url}"
      port_checksum=$(curl -s --fail "${port_url}" | sha256sum | cut -d' ' -f1)
      local error_hash="!e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
      if [ "${port_checksum}" = "${error_hash}" ]; then
        echo "Checksum FAILURE: port_checksum or url is not valid"
      else
        echo "Checksum SUCCESS: port_checksum set ${port_checksum}"
        break
      fi
    else
      echo "URL must end with .zip or .tar.gz. Please try again."
    fi
  done
  # TODO seperate build commands? shared libs?
  read_check 'Enter build command(s): ' port_build
  create_recipe "${port_folder}" "${port_url}" "${port_build}" "${port_checksum}" "${port_exe}"
  echo -e "\nCreated recipe for ${port_folder}_recipe.json\n"
  cat recipes/files/${port_folder}_recipe.json | jq '.'
  echo -e "\nCreated files/folders for ${port_folder}\n"
  tree new_ports/${port_folder}/
  echo -e "\nBefore disappearing to github. Please review the files created, add screenshots etc\n"
  echo -e "\nPlease refer to this guide for further information - https://portmaster.games/porting.html"
  echo -e "\nAfter making the changes do a git add, commit, push"
  echo -e "\nOpening github ${GITHUB_URL} to run workflow"
  xdg-open ${GITHUB_URL}  2>/dev/null

  exit_nicely
}

cd ..
clear
green='\033[32m'
nocolour='\033[0m'
port_date=$(date +%Y-%m-%d)

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
  jq -r --arg search "${port_title}" '
    .ports[]
    | select(.attr.title? != null and (.attr.title | test($search; "il")))
    | .attr.title
  ' releases/ports.json
)

if [ ${#ports[@]} -eq 0 ]; then
    echo -e "\n\nNo ports found with ${port_title} in title"
    echo "Please add basic port information - See the wiki https://github.com/monkeyx-net/PortMaster-Build-Templates/wiki"
    create_new_port "${port_title}"
fi

echo "Available ports with ${port_title} in the title:"
for i in "${!ports[@]}"; do
    echo "$((i + 1)) ${ports[${i}]}"
done

while true; do
    echo -n "Enter choice: "
    read choice
    if [[ ${choice} =~ ^[0-9]+$ ]] && [ ${choice} -ge 1 ] && [ ${choice} -le ${#ports[@]} ]; then
        port_title="${ports[$((choice-1))]}"
        break
    else
        echo "Invalid selection! Please enter a number between 1 and ${#ports[@]}"
    fi
done

echo "Now processing: ${port_title}"

exit_nicely

