import camelcaseKeys from "camelcase-keys";
import { coolcar } from "./services/gen/trip_pb";

// app.ts
App({
  globalData: {},
  onLaunch() {
    wx.request({
      url: "http://192.168.56.1:8080/trip/trip456",
      method: "GET",
      success(res) {
        const getTripRes = coolcar.GetTripResponse.fromObject(
          camelcaseKeys(res.data as Object, { deep: true })
        );
        console.log(getTripRes);
      },
      fail: console.error,
    });
    // 登录
    wx.login({
      success: (res) => {
        console.log(res.code);
        // 发送 res.code 到后台换取 openId, sessionKey, unionId
      },
    });
  },
});
