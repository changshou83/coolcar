import { IAppOption } from "../../types/appOptions";
import { routing } from "../../utils/util";

const enableShareKey = "enable_share_location";

Page({
  /* 页面状态 */
  carID: "",
  carRefresher: 0,
  /* 页面数据 */
  data: {
    avatarURL: "",
    enableShareLoc: false,
  },
  /* 生命周期函数 */
  async onLoad(opts: Record<"car_id", string>) {
    const { car_id } = opts as routing.LockOpts;
    this.carID = car_id;

    const userInfo = await getApp<IAppOption>().globalData.userInfo;
    this.setData({
      avatarURL: userInfo.avatarUrl,
      enableShareLoc: wx.getStorageSync(enableShareKey) || false,
    });
  },
  onUnload() {
    this.clearCarRefresher();
    wx.hideLoading();
  },
  /* 页面方法 */
  getUserInfo(evt: any) {
    const userInfo = evt.detail.userInfo as WechatMiniprogram.UserInfo;
    if (userInfo) {
      getApp<IAppOption>().resolveUserInfo(userInfo);
      wx.setStorageSync(enableShareKey, true);
      this.setData({
        enableShareLoc: true,
      });
    }
  },
  shareLocation(evt: any) {
    this.data.enableShareLoc = evt.detail.value;
    wx.setStorageSync(enableShareKey, this.data.enableShareLoc);
  },
  unlock() {
    wx.getLocation({
      type: "gcj02",
      success: async (location) => {
        if (!this.carID) {
          console.error("no car id specified");
          return;
        }
      },
      fail() {
        wx.showToast({
          icon: "none",
          title: "请前往设置页授权位置信息",
        });
      },
    });
  },
  /* 辅助方法 */
  clearCarRefresher() {
    if (this.carRefresher) {
      clearInterval(this.carRefresher);
      this.carRefresher = 0;
    }
  },
});
