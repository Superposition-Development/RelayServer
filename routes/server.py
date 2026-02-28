from flask import Blueprint, request, jsonify
import database
from AuthKeyGen import requiresToken
import time

bpServer = Blueprint("server",__name__)

def userInServer(userID,serverID):
    servers = database.queryTableValue("serverID","serverUser","userID",userID,True)
    return any(row[0] == serverID for row in servers)


def createServerUser(serverID,userID):
    serverUser = {
        "serverID": serverID,
        "userID": userID,
        "timestamp": int(time.time())
    }
    database.addRowAndReturnRowID(serverUser,"serverUser")


@bpServer.route("/createServer",methods=["POST"])
@requiresToken
def createServer(user):
    """
    Create a server in the server table 

    The expected payload from the client is as follows, 
    
    {
        name: the name of the server to add
        pfp: Base64 of picture
    }

    """
    data = request.get_json()

    createdServer = {
        "name": data["name"],
        "pfp": data["pfp"],
        "timestamp": int(time.time())
    }
    serverID = database.addRowAndReturnRowID(createdServer,"server")
    createServerUser(serverID,user["userID"])

    response = jsonify({
            "Message":"Server Created Successfully"
        })
    return response

# get servers a user is in, i admit this is not a great endpoint name but i dont want a roblox looking method name thats 80 words long
@bpServer.route("/getServers",methods=["GET"])
@requiresToken
def getServers(user):

    servers = database.queryTableValue("serverID","serverUser","userID",user["userID"],True)
    if(not servers):
        response = jsonify({
            "Message":"No servers found"
        })

    serverList = []
    print(servers)

    for i in servers:
        print(i[0])
        serverQuery = database.queryTableValue(["name","pfp"],"server","id",i[0])
        serverList.append({"name":serverQuery["name"],"pfp":serverQuery["pfp"],"id":i})

    response = jsonify({
            "Message":serverList
        })
    return response
