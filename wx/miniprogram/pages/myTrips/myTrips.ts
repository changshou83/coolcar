import { getProfile } from "../../apis/profile";
import { rental } from "../../apis/proto_gen/rental/rental_pb";
import { getTrips } from "../../apis/trip";
import { defaultAvatar } from "../../constants/index";
import { getUserInfo, routing, setUserInfo } from "../../utils/index";

interface Trip {
  id: string;
  start: string;
  end: string;
  duration: string;
  fee: string;
  distance: string;
  status: string;
}
// 视图列表元素
interface DayItem {
  id: string;
  tripItemID: string;
  label: string;
}

interface TripItem {
  id: string;
  dayItemID: string;
  data: Trip;
}
// DOM查询
interface TripItemQueryResult {
  id: string;
  top: number;
  dataset: {
    dayItemId: string;
  };
}

interface DayItemQueryResult {
  id: string;
}
// 滚动状态
interface ScrollState {
  tripItemID: string;
  dayItemID: string;
}

// const tripStatusMap = new Map([
//   [rental.v1.TripStatus.IN_PROGRESS, "进行中"],
//   [rental.v1.TripStatus.FINISHED, "已完成"],
// ]);

const IdentityStatus = rental.v1.IdentityStatus;
const licStatusMap = new Map([
  [IdentityStatus.UNSUBMITTED, "未认证"],
  [IdentityStatus.PENDING, "未认证"],
  [IdentityStatus.VERIFIED, "已认证"],
]);

Page({
  /* 页面状态 */
  tripListState: [] as TripItemQueryResult[],
  dayListState: [] as DayItemQueryResult[],
  /* 页面数据 */
  data: {
    licenseState: licStatusMap.get(IdentityStatus.UNSUBMITTED),
    avatarURL: "",
    trips: [] as TripItem[],
    tripScrollView: "",
    days: [] as DayItem[],
    selectedDay: "",
    dayListScrollTop: 0,
  },
  /* 生命周期函数 */
  async onLoad() {
    // 获取头像
    const { avatarURL = defaultAvatar } = getUserInfo();
    this.setData({
      avatarURL,
    });
  },
  async onShow() {
    const { status } = await getProfile();
    this.setData({
      licenseState: licStatusMap.get(status ?? IdentityStatus.UNSUBMITTED),
    });
  },
  async onReady() {
    await getTrips();
    // const { trips } = await getTrips();
    this.populateTrips();
  },
  /* 页面方法 */
  gotoRegister() {
    const url = routing.register();
    console.log(url);
    wx.navigateTo({
      url,
    });
  },
  gotoDriving(evt: any) {
    if (!evt.currentTarget.dataset.tripInProgress) {
      return;
    }
    const tripID = evt.currentTarget.dataset.tripId;
    tripID &&
      wx.redirectTo({
        url: routing.driving({
          trip_id: tripID,
        }),
      });
  },
  chooseAvatar(evt: any) {
    const { avatarUrl: avatarURL } = evt.detail;
    const userInfo = getUserInfo();
    if (avatarURL) {
      this.setData({ avatarURL });
      setUserInfo({ ...userInfo, avatarURL });
    }
  },
  selectDay(evt: any) {
    const tripItemID = evt.currentTarget.dataset.tripItemId;
    const dayItemID = evt.currentTarget.id;
    this.updateScrollState({ tripItemID, dayItemID });
  },
  tripsScroll(evt: any) {
    const top = evt.currentTarget?.offsetTop + evt.detail?.scrollTop;
    if (top !== undefined) {
      const selItem = this.tripListState.find((item) => {
        // 在前一个滑过了 5/8 时，才选中第二个
        return item.top >= parseInt(top) - 100;
      });
      if (selItem) {
        const dayItemID = selItem.dataset.dayItemId;
        this.updateScrollState({ tripItemID: "", dayItemID });
      }
    }
  },
  /* 辅助方法 */
  setupListState() {
    wx.createSelectorQuery()
      .selectAll(".trip")
      .fields({
        id: true,
        dataset: true,
        rect: true,
      })
      .exec(([tripList]) => {
        this.tripListState = tripList;
      });
    wx.createSelectorQuery()
      .selectAll(".day")
      .fields({ id: true })
      .exec(([dayList]) => {
        this.dayListState = dayList;
      });
  },
  populateTrips() {
    const days: DayItem[] = [];
    const trips: TripItem[] = [];
    let selectedDay;
    // let prevTripDate;
    for (let i = 0; i < 100; i++) {
      const dayItemID = "day-item-" + i;
      const tripItemID = "trip-item-" + i;

      const tripData: Trip = {
        id: (10001 + i).toString(),
        start: "东方明珠",
        end: "迪士尼",
        distance: "27.0",
        duration: "0时44分",
        fee: "128",
        status: "已完成",
      };

      // TODO:如果他们是同一天的，就不推新的day，并且新的trip的dayItemID与上一个相同
      // if(prevTripDate == currentTripDate) {}
      trips.push({ dayItemID, id: tripItemID, data: tripData });
      days.push({ tripItemID, id: dayItemID, label: (10001 + i).toString() });

      if (i === 0) {
        selectedDay = dayItemID;
      }
    }
    this.setData({ trips, days, selectedDay }, this.setupListState);
  },
  updateScrollState(state: ScrollState) {
    const { tripItemID, dayItemID } = state;
    const idx = this.dayListState.findIndex((item) => item.id === dayItemID);

    if (tripItemID !== undefined && dayItemID && idx !== -1) {
      this.setData({
        tripScrollView: tripItemID,
        selectedDay: dayItemID,
        dayListScrollTop: 45 * (idx === 0 ? idx : idx - 1),
      });
    }
  },
});
