#!/bin/sh
version=`cat VERSION`
target_name=BayServer_Go_${version}

version_file=`find . -name "version.go"`
sed  "s/\(const VERSION =\).*/\1 \"$version\"/"  $version_file > $version_file.bak
mv $version_file.bak $version_file

root=`dirname $0`
PLATFORMS=("windows/amd64" "linux/amd64" "darwin/amd64" "linux/arm" "darwin/arm")


build_for_os() {
    local platform=$1

    pushd .
    IFS="/" read -r os arch <<< "${platform}"
    output_name="${target_name}-${os}-${arch}"

    echo "***********************************************************"
    echo "      Building archive os=$os architecture=$arch"
    echo "***********************************************************"
    output_dir=/tmp/${output_name}
    rm -fr $output_dir

    mkdir $output_dir
    mkdir $output_dir/bin

    echo "Compiling Go files"
    pushd .
    cd modules/bayserver/main

    if [ "${os}" = "windows" ]; then
       bin_name=bayserver.exe
    else
       bin_name=bayserver
    fi

    GOOS=$os GOARCH=$arc go build -o $output_dir/bin/${bin_name}
#    GOOS=$os GOARCH=$arc go build -gcflags "all-=N -l" -o $output_dir/bin/bayserver.dbg
    popd

    echo "Copying files"
    mkdir $output_dir/plan
    cp -r stage/* $output_dir

    cd /tmp
    tar czf ${output_name}.tgz ${output_name}

    popd
}


for pf in "${PLATFORMS[@]}"; do
    build_for_os $pf
done
