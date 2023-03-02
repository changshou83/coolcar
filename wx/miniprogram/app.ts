// import camelcaseKeys from "camelcase-keys";
// import { auth } from "./apis/proto_gen/auth/auth_pb";
// import { rental } from "./apis/proto_gen/rental/rental_pb";
import { login } from "./utils/index";

// app.ts
App({
  globalData: {},
  onLaunch() {
    // 登录
    login();
  },
});
