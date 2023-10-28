#!/bin/sh

# create team
teamId=$(curl -sX POST http://localhost:3333/teams -d '{"name":"team 1"}' | jq .id)
[ -z "${teamId}" ] && {
	echo failed to create team
	exit 1
}
echo "created team ${teamId}"

# get team
teamName=$(curl -sX GET http://localhost:3333/teams/${teamId} | jq .name)
[ "${teamName}" != '"team 1"' ] && {
	echo failed to get team
	exit 1
}
echo "got team ${teamId} ${teamName}"

# update team
teamName=$(curl -sX PUT http://localhost:3333/teams/${teamId} -d '{"name":"team 2"}' | jq .name)
[ "${teamName}" != '"team 2"' ] && {
	echo failed to update team
	exit 1
}
echo "updated team ${teamId} ${teamName}"

# delete team
status=$(curl -isX DELETE http://localhost:3333/teams/${teamId} | grep "HTTP/1.1" | cut -d' ' -f 2)
[ "${status}" != '200' ] && {
	echo failed to delete team
	exit 1
}
echo "deleted team ${teamId}"

# create team
teamIdx=$(curl -sX POST http://localhost:3333/teams -d '{"name":"team x"}' | jq .id)
[ -z "${teamIdx}" ] && {
	echo failed to create team
	exit 1
}
echo "created team ${teamIdx}"

# create team
teamIdy=$(curl -sX POST http://localhost:3333/teams -d '{"name":"team y"}' | jq .id)
[ -z "${teamIdy}" ] && {
	echo failed to create team
	exit 1
}
echo "created team ${teamIdy}"


# list teams
numTeams=$(curl -sX GET http://localhost:3333/teams | jq .teams | jq length)
[ "${numTeams}" -lt 2 ] && {
	echo failed to list teams
	exit 1
}
echo "listd team num: ${numTeams}"


# create merchant
merchantId=$(curl -sX POST http://localhost:3333/merchants -d "{\"name\":\"merchant 1\",\"teamId\":${teamIdx}}" | jq .id)
[ -z "${merchantId}" ] && {
	echo failed to create merchant
	exit 1
}
echo "created merchant ${merchantId}"

# get merchant
merchantName=$(curl -sX GET http://localhost:3333/merchants/${merchantId} | jq .name)
[ "${merchantName}" != '"merchant 1"' ] && {
	echo failed to get merchant
	exit 1
}
echo "got merchant ${merchantId} ${merchantName}"

# update merchant
merchantName=$(curl -sX PUT http://localhost:3333/merchants/${merchantId} -d '{"name":"merchant 2"}' | jq .name)
[ "${merchantName}" != '"merchant 2"' ] && {
	echo failed to update merchant
	exit 1
}
echo "updated merchant ${merchantId} ${merchantName}"

# delete merchant
status=$(curl -isX DELETE http://localhost:3333/merchants/${merchantId} | grep "HTTP/1.1" | cut -d' ' -f 2)
[ "${status}" != '200' ] && {
	echo failed to delete merchant
	exit 1
}
echo "deleted merchant ${merchantId}"


# create merchant
merchantId=$(curl -sX POST http://localhost:3333/merchants -d "{\"name\":\"merchant x1\",\"teamId\":${teamIdx}}" | jq .id)
[ -z "${merchantId}" ] && {
	echo failed to create merchant
	exit 1
}
echo "created merchant ${merchantId}"

merchantId=$(curl -sX POST http://localhost:3333/merchants -d "{\"name\":\"merchant x2\",\"teamId\":${teamIdx}}" | jq .id)
[ -z "${merchantId}" ] && {
	echo failed to create merchant
	exit 1
}
echo "created merchant ${merchantId}"

merchantId=$(curl -sX POST http://localhost:3333/merchants -d "{\"name\":\"merchant y\",\"teamId\":${teamIdy}}" | jq .id)
[ -z "${merchantId}" ] && {
	echo failed to create merchant
	exit 1
}
echo "created merchant ${merchantId}"

#list merchants by team
numMerchantsX=$(curl -sX GET http://localhost:3333/teams/${teamIdx}/merchants | jq .merchants | jq length)
[ "${numMerchantsX}" -ne 2 ] && {
	echo failed to list merchant by team ${teamIdx}
	exit 1
}
echo "listed team num: ${numMerchantsX}"

#list merchants by team
numMerchantsY=$(curl -sX GET http://localhost:3333/teams/${teamIdy}/merchants | jq .merchants| jq length)
[ "${numMerchantsY}" -ne 1 ] && {
	echo failed to list merchant by team ${teamIdy}
	exit 1
}
echo "listed team num: ${numMerchantsY}"

# get all teams
ids=$(curl -sX GET http://localhost:3333/teams | jq .teams[].id)
for id in ${ids} ; do
	status=$(curl -isX DELETE http://localhost:3333/teams/${id} | grep "HTTP/1.1" | cut -d' ' -f 2)
	[ "${status}" != '200' ] && {
		echo failed to delete team ${id}
		exit 1
	}
	echo deleted team ${id}
done
numTeams=$(curl -sX GET http://localhost:3333/teams | jq .teams | jq length)
[ "${numTeams}" -ne 0 ] && {
	echo failed to delete all teams
	exit 1
}
echo "deleted all teams"

numMerchants=$(curl -sX GET http://localhost:3333/merchants| jq .merchants | jq length)
[ "${numMerchants}" -ne 0 ] && {
	echo failed to delete all merchants
	exit 1
}
echo "deleted all merchants"
