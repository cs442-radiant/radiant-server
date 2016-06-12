#!/usr/bin/env python

import argparse
from subprocess import call

parser = argparse.ArgumentParser()

parser.add_argument("-p", "--port", type=int, default=8100)
parser.add_argument("--host", default="ec2-54-191-70-38.us-west-2.compute.amazonaws.com")
parser.add_argument("command")
parser.add_argument("-l", "--local", action='store_true')
parser.add_argument("-d", "--data", help="Path to .json file.")

args = parser.parse_args()

curlArgs = ["curl", "-i", "-v"]

host = args.host
port = args.port

isValidCommand = False
commandUri = ""
method = ""

if args.command == "PostBundle":
    isValidCommand = True
    commandUri = "bundle"
    method = "POST"
elif args.command == "PostSample":
    isValidCommand = True
    commandUri = "sample"
    method = "POST"
elif args.command == "PostLearn":
    isValidCommand = True
    commandUri = "learn"
    method = "POST"

if method != "":
    curlArgs.append("-X")
    curlArgs.append(method)

if args.data != "" and not (args.data is None):
    curlArgs.append("-H")
    curlArgs.append("Content-Type: application/json")
    curlArgs.append("--data-binary")
    curlArgs.append("@%s" % args.data)

if args.local:
    host = "localhost"

fullUri = "%s:%s/%s" % (host, port, commandUri)
curlArgs.append(fullUri)

if isValidCommand:
    print(" ".join(curlArgs))
    call(curlArgs)
else:
    print "Not a valid command."