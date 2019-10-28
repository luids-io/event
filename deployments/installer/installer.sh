#!/bin/bash

## Configuration variables. 
RELEASE="v0.0.1"
ARCH="amd64"
SVC_USER=luevent
ETC_DIR=/etc/luevent
BIN_DIR=/usr/local/bin
VAR_DIR=/var/lib/luevent
CACHE_DIR=/var/cache/luevent
SYSTEMD_DIR=/etc/systemd/system
DOWNLOAD_BASE="https://github.com/luids-io/event/releases/download"
DOWNLOAD_URI="${DOWNLOAD_BASE}/${RELEASE}/luevent_${RELEASE}_linux_${ARCH}.tgz"
##

die() { echo "error: $@" 1>&2 ; exit 1 ; }

## some checks
for deps in "wget" "mktemp" "getent" "useradd" ; do
	which $deps >/dev/null \
		|| die "$deps is required!"
done
[[ $EUID -eq 0 ]] || die "This script must be run as root"
[ -d $BIN_DIR ] || die "Binary directory $BIN_DIR doesn't exist"

## options command line
OPT_UNATTEND=0
OPT_OVERWRITE_BIN=0
while [ -n "$1" ]; do
	case "$1" in
		-u) OPT_UNATTEND=1 ;;
		-o) OPT_OVERWRITE_BIN=1 ;; 
		-h) echo -e "Options:\n\t [-u] unattend\n\t [-o] overwrite binaries\n"
		    exit 0 ;; 
 		*) die "Option $1 not recognized" ;; 
	esac
	shift
done

echo
echo "================="
echo "luEvent installer "
echo "================="
echo

show_actions() {
	echo "Warning! This script will commit the following changes to your system:"
	echo ". Download and install binaries in '${BIN_DIR}'"
	echo ". Create a system user '${SVC_USER}'"
	echo ". Create data dir '${VAR_DIR}'"
	echo ". Create cache dir '${CACHE_DIR}'"
	echo ". Create config dir '${ETC_DIR}'"
	[ -d $SYSTEMD_DIR ] && echo ". Copy systemd configurations to '${SYSTEMD_DIR}'"
	echo ""
}

if [ $OPT_UNATTEND -eq 0 ]; then
	show_actions
	read -p "Are you sure? (y/n) " -n 1 -r
	echo
	echo
	if [[ ! $REPLY =~ ^[Yy]$ ]]
	then
		die "canceled"
	fi
fi

TMP_DIR=$(mktemp -d -t ins-XXXXXX) || die "couldn't create temp"
LOG_FILE=${TMP_DIR}/installer.log

log() { echo `date +%y%m%d%H%M%S`": $@" >>$LOG_FILE ; }
step() { echo -n "* $@..." ; log "STEP: $@" ; }
step_ok() { echo " OK" ; }
step_err() { echo " ERROR" ; }
user_exists() { getent passwd $1>/dev/null ; }
group_exists() { getent group $1>/dev/null ; }

## do functions
do_download() {
	[ $# -eq 2 ] || die "${FUNCNAME}: unexpected number of params"
	local url="$1"
	local filename="$2"

	local dst="${TMP_DIR}/${filename}"
	rm -f $dst
	log "downloading $url"
	echo "$url" | grep -q "^\(http\|ftp\)"
	if [ $? -eq 0 ]; then
		wget "$url" -O $dst &>>$LOG_FILE
	else
		cp -v "$url" $dst &>>$LOG_FILE
	fi
}

do_clean_file() {
	[ $# -eq 1 ] || die "${FUNCNAME}: unexpected number of params"
	local filename=$1

	local src="${TMP_DIR}/${filename}"
	log "clearing $src"    
	rm -f $src &>>$LOG_FILE
}

do_install_bin() {
	[ $# -eq 1 ] || die "${FUNCNAME}: unexpected number of params"
	local binary=$1

	local src="${TMP_DIR}/${binary}"
	local dst="${BIN_DIR}/${binary}"
	[ ! -f $src ] && log "$src not found!" && return 1

	log "copying $src to $dst, chown root, chmod 755"
	{ cp $src $dst \
		&& chown root:root $dst \
		&& chmod 755 $dst
	} &>>$LOG_FILE
}

do_unpackage() {
	[ $# -eq 1 ] || die "${FUNCNAME}: unexpected number of params"
	local tgzfile=$1
	
	local src="${TMP_DIR}/${tgzfile}"
	[ ! -f $src ] && log "${FUNCNAME}: $src not found!" && return 1

	log "unpackaging $tgzfile"
	tar -zxvf $src -C $TMP_DIR &>>$LOG_FILE
}

do_create_datadir() {
	[ $# -ge 2 ] || die "${FUNCNAME}: unexpected number of params"
	local datadir=$1
	local datagrp=$2
	local perm=1775
	if [ $# -gt 2 ]; then
		perm=$3
	fi

	[ -d $datadir ] && log "$datadir found!" && return 1
	group_exists $datagrp || { log "group $datagrp doesn't exists" && return 1 ; }

	log "creating dir $datadir, chgrp to $datagrp, chmod $perm"
	{ mkdir -p $datadir \
		&& chown root:$datagrp $datadir \
		&& chmod $perm $datadir
	} &>>$LOG_FILE
}

do_create_sysuser() {
	[ $# -eq 2 ] || die "${FUNCNAME}: unexpected number of params"
	local nuser=$1
	local nhome=$2

	user_exists $nuser && log "user $nuser already exists" && return 1

	log "useradd $nuser with params"
	useradd -s /usr/sbin/nologin -r -M -d $nhome $nuser &>>$LOG_FILE
}

## steps
install_binaries() {
	step "Downloading and installing binaries"
	if [ $OPT_OVERWRITE_BIN -eq 0 ]; then
		[ -f ${BIN_DIR}/eventproc ] \
			&& log "${BIN_DIR}/eventproc already exists" \
			&& step_ok && return 0
	fi
	do_download "$DOWNLOAD_URI" luevent_linux.tgz
	[ $? -ne 0 ] && step_err && return 1

	do_unpackage luevent_linux.tgz
	[ $? -ne 0 ] && step_err && return 1
	do_clean_file luevent_linux.tgz

	for binary in "eventproc" "eventnotify" ; do
		do_install_bin $binary
		[ $? -ne 0 ] && step_err && return 1
        	do_clean_file $binary
	done

	step_ok
}

create_system_user() {
	step "Creating system user"
	user_exists $SVC_USER \
		&& log "user $SVC_USER already exists" && step_ok && return 0
	
	do_create_sysuser "$SVC_USER" "$VAR_DIR"
	[ $? -ne 0 ] && step_err && return 1
	
	step_ok
}

create_data_dir() {
	step "Creating data dir"
	[ -d $VAR_DIR ] && log "$VAR_DIR already exists" && step_ok && return 0

	do_create_datadir $VAR_DIR $SVC_USER 1770
	[ $? -ne 0 ] && step_err && return 1

	step_ok
}

create_cache_dir() {
	step "Creating cache dir"
	[ -d $CACHE_DIR ] && log "$CACHE_DIR already exists" && step_ok && return 0

	do_create_datadir $CACHE_DIR $SVC_USER 1770
	[ $? -ne 0 ] && step_err && return 1

	step_ok
}

create_config() {
	step "Creating config dir with sample files"
	if [ ! -d $ETC_DIR ]; then
		log "creating dir $ETC_DIR"
		{ mkdir -p $ETC_DIR \
			&& chown root:root $ETC_DIR \
			&& chmod 755 $ETC_DIR
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1

		local ssldir="${ETC_DIR}/ssl"
		log "creating dir $ssldir with subdirs"
		{ mkdir -p ${ssldir}/certs  ${ssldir}/private \
			&& chown root:root ${ssldir}/certs \
			&& chmod 755 ${ssldir}/certs \
			&& chown root:$SVC_USER ${ssldir}/private \
			&& chmod 750 ${ssldir}/private
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$ETC_DIR already exists"
	fi

	if [ ! -f $ETC_DIR/eventproc.toml ]; then
		log "creating $ETC_DIR/eventproc.toml"
		{ cat > $ETC_DIR/eventproc.toml <<EOF
[eventproc]
dirs      = [ "${ETC_DIR}/stacks.d" ]

[stackbuild]
datadir   = "${VAR_DIR}"
cachedir  = "${CACHE_DIR}"

#[grpc-notify]
#listenuri  = "tcp://0.0.0.0:5851"
EOF
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$ETC_DIR/eventproc.toml already exists"
	fi

	local stacksd=${ETC_DIR}/stacks.d
	if [ ! -d $stacksd ]; then
		log "creating dir $stacksd"
		{ mkdir -p $stacksd \
			&& chown root:root $stacksd \
			&& chmod 755 $stacksd
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$stacksd already exists"
	fi
	if [ ! -f $stacksd/main.json ]; then
		log "creating $stacksd/main.json"
		{ cat > $stacksd/main.json <<EOF
[
  {
    "name": "main",
    "modules": [
      {
        "name": "archive",
        "filters": [
          { "class": "basicexpr",
            "args": [ "type", "==", "security" ] }
        ],
        "plugins": [
          { "class": "jsonwriter",
            "args": [ "security-events.json" ] }
        ],
        "onsuccess": "next",
        "onerror": "stop"
      }
    ]
  }
]
EOF
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$stacksd/main.json already exists" && step_ok && return 0
	fi

	step_ok
}

install_systemd_services() {
	step "Installing systemd services"
	if [ ! -f $SYSTEMD_DIR/eventproc.service ]; then
		log "creating $SYSTEMD_DIR/eventproc.service"
		{ cat > $SYSTEMD_DIR/eventproc.service <<EOF
[Unit]
Description=eventproc service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=on-failure
RestartSec=1
User=$SVC_USER
ExecStart=$BIN_DIR/eventproc --config $ETC_DIR/eventproc.toml

[Install]
WantedBy=multi-user.target
EOF
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$SYSTEMD_DIR/eventproc.service already exists"
	fi

	if [ ! -f $SYSTEMD_DIR/eventproc@.service ]; then
		log "creating $SYSTEMD_DIR/eventproc@.service"
		{ cat > $SYSTEMD_DIR/eventproc@.service <<EOF
[Unit]
Description=eventproc service per-config file
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=on-failure
RestartSec=1
User=$SVC_USER
ExecStart=$BIN_DIR/eventproc--config $ETC_DIR/%i.toml

[Install]
WantedBy=multi-user.target
EOF
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$SYSTEMD_DIR/eventproc@.service already exists"
	fi

	step_ok
}

## main process

install_binaries || die "Show $LOG_FILE"
create_system_user || die "Show $LOG_FILE"
create_data_dir || die "Show $LOG_FILE"
create_cache_dir || die "Show $LOG_FILE"
create_config || die "Show $LOG_FILE"
[ -d $SYSTEMD_DIR ] && { install_systemd_services || die "Show $LOG_FILE for details." ; }

echo
echo "Installation success!. You can see $LOG_FILE for details."
