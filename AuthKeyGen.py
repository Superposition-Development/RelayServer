import jwt
import datetime
import init
from flask import jsonify, request
from functools import wraps
import database

secretKey = init.SECRET_KEY

def encryptJWT(payload,expireTimeMin:int):
    payload["exp"] = datetime.datetime.utcnow() + datetime.timedelta(minutes=expireTimeMin)
    return jwt.encode(payload=payload,key=secretKey,algorithm="HS256")

def decryptJWT(token):
    status = {"Error":"JWT decryption failed for unknown reason"}
    try:
        status = {"Result":jwt.decode(jwt=token,key=secretKey,algorithms=["HS256"])}
    except jwt.ExpiredSignatureError:
        status = {"Error":"JWT expired"}
    except jwt.InvalidSignatureError:
        status = {"Error":"https://tenor.com/view/i-know-gif-3951224689799851379"}
    print(status)
    return status

def requiresToken(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        token = request.headers.get("Authorization")
        if not token:
            return jsonify({'Error': 'Unauthorized Access, missing JWT'}), 401
        
        if token and token.startswith("Bearer "):
            token = token[7:]

        try:
            result = decryptJWT(token)
            if("Error" in result.keys()):
                return jsonify({'Error': 'Invalid JWT'}), 401
            
            user = database.queryTableValue(["id","pfp","username","userID","password","timestamp"],"user","userID",result["Result"]["userID"])
            if(user == None):
                return jsonify({'Error': 'Invalid JWT'}), 401
    
        except Exception as e:
            return jsonify({'Error': e}), 401

        return f(user, *args, **kwargs)

    return decorated