#
# This software is distributed on an "AS IS" basis, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
# either express or implied, as more fully set forth in the License.
#
# See the NOTICE file distributed with this work for information regarding copyright ownership.
#

# This script is designed so that someone can update all versions for releases
# using a single command.

# Debugging
# set -x
set -e # exit if a command fails
set -u # unbound variable is an error

HELP=$(cat <<EOF
Usage: update-version.sh <old version> <new version>

    <old version>    The version to replace throughout the codebase
                     (i.e. 1.1.0-SNAPSHOT)
    <new version>    The new version in the codebase (i.e. 1.2.0-SNAPSHOT)
Be sure to escape the dots and quote the versions. For example,
update-version.sh "1\.1\.0-SNAPSHOT" "1\.2\.0-SNAPSHOT"
EOF
)

# Arguments:
#  $1: old version
#  $2: new version
function update_csi_version_string() {
  perl -pi -e "s/${1}/${2}/g" csi/alluxio/driver.go
}

# Arguments:
#  $1: old version
#  $2: new version
function update_dockerfiles() {
  perl -pi -e "s/${1}/${2}/g" dev/build/csi/Dockerfile
  perl -pi -e "s/${1}/${2}/g" dev/build/Dockerfile
}

# Arguments:
#  $1: old version
#  $2: new version
function update_chart_yaml() {
  find . -name Chart.yaml | xargs -t -n 1 perl -pi -e "s/${1}/${2}/g"
}

# Arguments:
#  $1: old version
#  $2: new version
function update_tests() {
  find tests/helm/expectedTemplates -name *.yaml | xargs -t -n 1 perl -pi -e "s/${1}/${2}/g"
}

function main() {
    if [[ "$#" -ne 2 ]]; then
        echo "Arguments '<old version>' and '<new version>' must be provided."
        echo "${HELP}"
        exit 1
    fi

    local _old="${1}"
    local _new="${2}"

    update_csi_version_string "$_old" "$_new"
    update_dockerfiles "$_old" "$_new"
    update_chart_yaml "$_old" "$_new"
    update_tests "$_old" "$_new"

    exit 0
}

main "$@"
