from flask import Blueprint, request, jsonify
import database
from AuthKeyGen import requiresToken
import time
from routes.server import userInServer

bpChannel = Blueprint("channel",__name__)

@bpChannel.route("/createChannel",methods=["POST"])
@requiresToken
def createChannel(user):
    """
  write smth here

    """
    data = request.get_json()
    if not(userInServer(user["userID"],data["serverID"])):
        response = jsonify({
            "Error":"You are not in that serer"
        })

    createdChannel = {
        "name": data["name"],
        "serverID": data["serverID"],
    }
    
    database.addRowAndReturnRowID(createdChannel,"channel")

    response = jsonify({
            "Message":"Channel Created Successfully"
        })
    return response

# get servers a user is in, i admit this is not a great endpoint name but i dont want a roblox looking method name thats 80 words long
@bpChannel.route("/getChannels",methods=["POST"])
@requiresToken
def getChannels(user):

    data = request.get_json()
    if not(userInServer(user["userID"],data["serverID"])):
        response = jsonify({
            "Error":"You are not in that serer"
        })

    channels = database.queryTableValue(["name","id"],"channel","serverID",data["serverID"],True)
    if(not channels):
        response = jsonify({
            "Message":"No channels found"
        })

    channelList = []

    for channel in channels.values():
        channelList.append({"name":channel[0],"id":channel[1]})

    response = jsonify({
            "Message":channelList
        })
    return response
