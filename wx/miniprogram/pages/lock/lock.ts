import { getCar } from "../../apis/car";
import { car } from "../../apis/proto_gen/car/car_pb";
import { createTrip } from "../../apis/trip";
import { defaultAvatar } from "../../constants/index";
import { getUserInfo, routing, setUserInfo } from "../../utils/index";

const enableShareKey = "enable_share_location";

Page({
  /* 页面状态 */
  carID: "",
  carRefresher: undefined as number | undefined,
  /* 页面数据 */
  data: {
    avatarUrl: defaultAvatar,
    enableShareLoc: false,
  },
  /* 生命周期函数 */
  async onLoad() {
    // async onLoad(opts: Record<"car_id", string>) {
    // const { car_id }: routing.LockOpts = opts;
    // this.carID = car_id;
    this.carID = "645ae792ffdf429ae3b1439c";

    const { avatarURL } = getUserInfo();
    this.setData({
      avatarUrl: avatarURL || defaultAvatar,
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
      success: async (loc) => {
        if (!this.carID) {
          console.error("no car id specified");
          return;
        }
        // create trip
        const { id } = await createTrip({
          start: {
            latitude: loc.latitude,
            longitude: loc.longitude,
          },
          carId: this.carID,
          avatarUrl: this.data.avatarUrl,
        });
        // show toast
        wx.showToast({
          title: "开锁中...",
          mask: true,
        });
        // create car refresher
        this.carRefresher = setInterval(async () => {
          const c = await getCar(this.carID);
          if (c.status === car.v1.CarStatus.UNLOCKED) {
            this.clearCarRefresher();
            wx.redirectTo({
              url: routing.driving({ trip_id: id }),
              complete() {
                wx.hideLoading();
              },
            });
          }
        }, 2000);
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
    const userInfo = getUserInfo();
    if (avatarUrl) {
      this.setData({ avatarUrl });
      setUserInfo({ ...userInfo, avatarURL: avatarUrl });
    }
  },
  /* 辅助方法 */
  clearCarRefresher() {
    if (this.carRefresher) {
      clearInterval(this.carRefresher);
      this.carRefresher = undefined;
    }
  },
});
