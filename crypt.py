import jwt
import datetime
import init

jwtKey = init.JWT_KEY

def encryptJWT(payload,expireTimeMin:int):
    payload["exp"] = datetime.datetime.now() + datetime.timedelta(minutes=expireTimeMin)
    return jwt.encode(payload=payload,key=jwtKey,algorithm="HS256")

def decryptJWT(token):
    status = {"Error":"JWT decryption failed for unknown reason"}
    try:
        status = {"Result":jwt.decode(jwt=token,key=jwtKey,algorithms=["HS256"])}
    except jwt.ExpiredSignatureError:
        status = {"Error":"JWT expired"}
    return status