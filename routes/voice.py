from flask import Blueprint, request, jsonify
import database
from AuthKeyGen import requiresToken
import time
from routes.server import userInServer
from init import socketApp

bpVoice = Blueprint("voice",__name__)

