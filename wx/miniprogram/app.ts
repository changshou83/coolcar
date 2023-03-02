import camelcaseKeys from "camelcase-keys";
import { auth } from "./services/proto_gen/auth/auth_pb";
import { rental } from "./services/proto_gen/rental/rental_pb";

// app.ts
App({
  globalData: {},
  onLaunch() {
    // wx.request({
    //   url: "http://192.168.56.1:8080/trip/trip456",
    //   method: "GET",
    //   success(res) {
    //     const getTripRes = coolcar.GetTripResponse.fromObject(
    //       camelcaseKeys(res.data as Object, { deep: true })
    //     );
    //     console.log(getTripRes);
    //   },
    //   fail: console.error,
    // });
    // 登录
    wx.login({
      success: (res) => {
        console.log(res.code);
        // 发送 res.code 到后台换取 openId, sessionKey, unionId
        wx.request({
          url: "http://192.168.56.1:8080/v1/auth/login",
          method: "POST",
          data: {
            code: res.code,
          },
          success(res) {
            const loginRes: auth.v1.ILoginResponse =
              auth.v1.LoginResponse.fromObject(
                camelcaseKeys(res.data as Object, { deep: true })
              );
            console.log(loginRes);
            wx.request({
              url: "http://localhost:8080/v1/trip",
              method: "POST",
              data: {
                start: "abc",
              } as rental.v1.ICreateTripRequest,
              // header: {
              //   authorization: "Bearer " + loginRes.assessToken,
              // },
            });
          },
          fail: console.error,
        });
      },
    });
  },
});
