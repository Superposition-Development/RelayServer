import jwt
import datetime
import init
from init import app
from flask import Flask, jsonify, request
from functools import wraps
import database

secretKey = init.SECRET_KEY

def encryptJWT(payload,expireTimeMin:int):
    payload["exp"] = datetime.datetime.now() + datetime.timedelta(minutes=expireTimeMin)
    return jwt.encode(payload=payload,key=secretKey,algorithm="HS256")

def decryptJWT(token):
    status = {"Error":"JWT decryption failed for unknown reason"}
    try:
        status = {"Result":jwt.decode(jwt=token,key=secretKey,algorithms=["HS256"])}
    except jwt.ExpiredSignatureError:
        status = {"Error":"JWT expired"}
    return status

def token_required(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        token = request.cookies.get('JWTAuthKey')

        if not token:
            return jsonify({'message': 'Token is missing!'}), 401

        try:
            data = decryptJWT(token=token)
            if("Error" in data.keys()):
                pass #this should not be pass i am just really not sure what to do

            if(data["userID"]):
                pass

            if(database.queryUser()):
                pass

            # user = database.
            # current_user = User.query.filter_by(public_id=data['public_id']).first()
        except:
            return jsonify({'message': 'Token is invalid!'}), 401

        return f("nothing", *args, **kwargs)

    return decorated