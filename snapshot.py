import requests
import json
import argparse
import sys

def shortShaToLongSha(username, repository, shortHash, token=""):
        resource_url = "https://api.github.com/search/commits?q=repo:{}/{}+hash:{}".format(username, repository, shortHash)
        headers = {"Accept": "application/vnd.github.cloak-preview+json"}
        if token != "":
            headers["Authorization"] = "token {}".format(token)

        r = requests.get(resource_url, headers=headers)

        if r.status_code != requests.codes.ok:
            raise ValueError("status code not ok: {}, body: {}".format(r.status_code, r.text ))

        data = r.json()

        if data["total_count"] != 1:
            raise ValueError("total_count not 1: {}".format(data["total_count"]))

        return data["items"][0]["sha"]

def versionRefToLongSha(username, repository, versionRef, token=""):
        resource_url = "https://api.github.com/repos/{}/{}/git/ref/tags/{}".format(username, repository, versionRef)
        headers = {"Accept": "application/vnd.github.cloak-preview+json"}
        if token != "":
            headers["Authorization"] = "token {}".format(token)

        r = requests.get(resource_url, headers=headers)

        if r.status_code != requests.codes.ok:
            raise ValueError("status code not ok: {}, body: {}".format(r.status_code, r.text ))

        data = r.json()
        return data["object"]["sha"]

def getOptions(args=sys.argv[1:]):
    parser = argparse.ArgumentParser(description="Parses command.")
    parser.add_argument("-t", "--token", type=str, help="Github personal token.")
    parser.add_argument("-m", "--modules", type=str, help="Path to vendor/modules.txt file.")
    parser.add_argument("-i", "--ip2ppmapping", type=str, help="Path to ip2pp mapping file.")
    parser.add_argument("-o", "--output", type=str, help="Target Godeps.json file.")
    parser.add_argument("-p", "--project", type=str, help="Project vendor/modules.txt file belongs to.")
    options = parser.parse_args(args)
    return options

if __name__ == "__main__":
    options = getOptions(sys.argv[1:])
    print(options)

    if options.modules == None:
        raise ValueError("--modules option not set")

    if options.ip2ppmapping == None:
        raise ValueError("--ip2ppmapping option not set")

    if options.project == None:
        raise ValueError("--project option not set")

    token = ""
    if options.token != None:
        token = options.token

    with open(options.ip2ppmapping) as data:
        ip2pp = json.loads(data.read())

    with open(options.modules) as data:
        lines = data.readlines()

    deps = []

    for line in lines:
        if not line.startswith("# "):
            continue
        line = line.strip()
        parts = line.split(" ")
        if len(parts) != 3:
            raise ValueError("line {} does not have 3 pieces, it has {}".format(line, len(parts)))

        userRepo = parts[1]
        newUserRepo = userRepo
        for ip in ip2pp:
            if userRepo.startswith(ip["ipprefix"]):
                newUserRepo = ip["provider_prefix"] + userRepo[len(ip["ipprefix"]):]
                break


        if newUserRepo.startswith("github.com/"):
            bits = newUserRepo.split("/")
            username = bits[1]
            repository = bits[2]
        else:
            raise ValueError("Repo {} not github".format(userRepo))

        verParts = parts[2].split("-")
        if len(verParts) == 1:
            sha = versionRefToLongSha(username, repository, verParts[0], token)
            deps.append({"ImportPath": userRepo, "Rev": sha})
            print("{}/{}, version: {}, Lhash: {}".format(username, repository, verParts[0], sha))
        elif len(verParts) == 3:
            sha = shortShaToLongSha(username, repository, verParts[2], token)
            deps.append({"ImportPath": userRepo, "Rev": sha})
            print("{}/{}, Shash: {}, Lhash: {}".format(username, repository, verParts[2], sha))

    godeps = {
    	"ImportPath": options.project,
    	"GoVersion": "go1.7",
    	"GodepVersion": "v74",
    	"Deps": deps
        }

    if options.output != None:
        with open(options.output, "w") as f:
            f.write(json.dumps(godeps))
    else:
        print(json.dumps(godeps))
