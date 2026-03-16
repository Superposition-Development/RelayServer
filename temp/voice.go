from flask import Blueprint, request, jsonify
import database
from AuthKeyGen import requiresToken
import time
from routes.server import userInServer

bpVoice = Blueprint("voice",__name__)

