import { formatFee, routing } from "../../utils/index";

// const updaterInterval = 5;
const initialLat = 42.05297;
const initialLng = 123.52658;

Page({
  /* 页面状态 */
  timer: undefined as number | undefined,
  tripID: "",
  /* 页面数据 */
  data: {
    scale: 12,
    fee: "0.00",
    elapsed: "00:00:00",
    location: {
      latitude: initialLat,
      longitude: initialLng,
    },
    markers: [
      {
        iconPath: "/resources/car.png",
        id: 0,
        latitude: initialLat,
        longitude: initialLng,
        width: 20,
        height: 20,
      },
    ],
  },
  /* 生命周期方法 */
  onLoad(opts: Record<"trip_id", string>) {
    const { trip_id } = opts as routing.DrivingOpts;
    this.tripID = trip_id;
    // setup
    this.setupLocationUpdater();
    this.setupTimer();
  },
  onUnload() {
    wx.stopLocationUpdate();
    this.timer && clearInterval(this.timer);
  },
  /* 页面方法 */
  async endTrip() {
    try {
      // await TripService.finishTrip(this.tripID);
      wx.redirectTo({
        url: "/pages/myTrips/myTrips",
      });
    } catch (err) {
      console.error(err);
      wx.showToast({
        title: "结束行程失败",
        icon: "none",
      });
    }
  },
  /* 辅助方法 */
  async setupTimer() {
    // const { start, current } = await TripService.getTrip(this.tripID);
    // let sinceLastUpdate = 0;
    // let lastUpdateDuration = current!.timestampSec! - start!.timestampSec!;
    // this.updateMarkers(current);
    // this.setData({
    //   elapsed: formatElapsed(sinceLastUpdate + lastUpdateDuration),
    // });

    this.timer = setInterval(() => {
      // sinceLastUpdate++;
      // if (sinceLastUpdate % updaterInterval === 0) {
      //   try {
      //     const { start, current } = await TripService.getTrip(this.tripID);
      //     sinceLastUpdate = 0;
      //     lastUpdateDuration = current!.timestampSec! - start!.timestampSec!;
      //     this.updateMarkers(current);
      //   } catch (err) {
      //     console.error(err);
      //   }
      // }
      // this.setData({
      //   elapsed: formatElapsed(sinceLastUpdate + lastUpdateDuration),
      // });
    }, 1000);
  },
  setupLocationUpdater() {
    wx.startLocationUpdate({
      fail: console.error,
    });
    wx.onLocationChange((loc) => {
      const location = {
        latitude: loc.latitude,
        longitude: loc.longitude,
      };
      this.setData({ location });
    });
  },
  updateMarkers(current: any) {
    const { markers } = this.data;
    const location = {
      latitude: current?.location?.latitude || initialLat,
      longitude: current?.location?.longitude || initialLng,
    };

    markers[0].latitude = location.latitude;
    markers[0].longitude = location.longitude;
    this.setData({
      location,
      fee: formatFee(current!.feeCent!),
      markers: markers,
    });
  },
});
