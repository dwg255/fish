namespace go rpc

enum ErrorCode{
    Success=0
    UnknownError=5000,
    VerifyError=5001,
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
    Result getUserInfoByken(1:string token)

    //修改用户金币
    Result modifyGoldById(1:string behavior, 2:i32 userId, 3:i64 gold)
    Result modifyGoldByToken(1:string behavior, 2:string token,3:i64 gold)
}
