import { routing } from "../../utils";

const enableShareKey = "enable_share_location";
const avatarURLKey = "avarat_URL";
const defaultAvatarUrl =
  "https://mmbiz.qpic.cn/mmbiz/icTdbqWNOwNRna42FI242Lcia07jQodd2FJGIYQfG0LAJGFxM4FbnQP6yfMxBgJ0F3YRqJCJ1aPAK2dQagdusBZg/0";

Page({
  /* 页面状态 */
  carID: "",
  carRefresher: 0,
  /* 页面数据 */
  data: {
    avatarUrl: defaultAvatarUrl,
    enableShareLoc: false,
  },
  /* 生命周期函数 */
  async onLoad(opts: Record<"car_id", string>) {
    const { car_id } = opts as routing.LockOpts;
    this.carID = car_id;

    this.setData({
      avatarUrl: wx.getStorageSync(avatarURLKey) || defaultAvatarUrl,
      enableShareLoc: wx.getStorageSync(enableShareKey) || false,
    });
  },
  onUnload() {
    this.clearCarRefresher();
    wx.hideLoading();
  },
  /* 页面方法 */
  shareLocation(evt: any) {
    const { value } = evt.detail;
    this.data.enableShareLoc = value; // 不用setData进行重新渲染
    wx.setStorageSync(enableShareKey, value);
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
  chooseAvatar(evt: any) {
    const { avatarUrl } = evt.detail;
    if (avatarUrl) {
      this.setData({ avatarUrl });
      wx.setStorageSync(avatarURLKey, avatarUrl);
    }
  },
  /* 辅助方法 */
  clearCarRefresher() {
    if (this.carRefresher) {
      clearInterval(this.carRefresher);
      this.carRefresher = 0;
    }
  },
});
