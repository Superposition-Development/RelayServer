from flask import Blueprint, request, jsonify
import database
from AuthKeyGen import requiresToken
import time
from routes.server import userInServer

bpMessage = Blueprint("message",__name__)

@bpMessage.route("/sendMessage",methods=["POST"])
@requiresToken
def sendMessage(user):
    """
    something here

    """
    data = request.get_json()

    if not(userInServer(user["userID"],data["serverID"])):
        response = jsonify({
            "Error":"You are not in that server"
        })

    createdMessage = {
        "senderID": user["userID"],
        "channelID":data["channelID"],
        "content":data["content"],
        "timestamp": int(time.time())
    }

    database.addRowAndReturnRowID(createdMessage,"message")

    response = jsonify({
            "Message":"Server Created Successfully"
        })
    return response

@bpMessage.route("/getMessage",methods=["POST"])
@requiresToken
def sendMessage(user):
    """
    something here

    """
    data = request.get_json()

    if not(userInServer(user["userID"],data["serverID"])):
        response = jsonify({
            "Error":"You are not in that server"
        })

    createdMessage = {
        "senderID": user["userID"],
        "channelID":data["channelID"],
        "content":data["content"],
        "timestamp": int(time.time())
    }

    database.addRowAndReturnRowID(createdMessage,"message")

    response = jsonify({
            "Message":"Server Created Successfully"
        })
    return response