namespace go rpc

enum ErrorCode{
    Success=0
    ServerError=5000,
    VerifyError=5001,
    UserNotExists=5002,
}

enum ModifyPropType{
    gems = 0
    roomId = 1
    power = 2
    ice = 3
}

struct UserInfo{
     1: i64 userId
     2: string userName
     3: string nickName
     4: i8 sex
     5: string headImg
     6: i32 lv
     7: i64 exp
     8: i8 vip
     9: i64 gems
     10: i64 roomId
     11: i64 power
     12: i8 reNameCount
     13: i8 reHeadCount
     14: string registerDate
     15: i64 ice
     16: string token
}

struct Result{
    1:  ErrorCode code
    2: UserInfo user_obj
}

service UserService {

    Result createNewUser(1: string nickName 2:string avatarAuto 3: i64 gold )//初始金币
    //获取用户信息 BY userId
    Result getUserInfoById(1:i32 userId)

    //获取用户信息 BY token
    Result getUserInfoByToken(1:string token)

    //修改用户金币
    Result modifyUserInfoById(1:string behavior, 2:i32 userId, 3: ModifyPropType propType, 4: i64 incr)
    Result RenameUserById(1:i32 userId,2:string NewName)
    string getMessage(1 :string messageType)
}
