import camelcaseKeys from "camelcase-keys";
import { wsURL } from "../constants/index";
import { car } from "./proto_gen/car/car_pb";
import { requestWithRetry } from "../utils/index";

export function subscribe(onMsg: (c: car.v1.ICarEntity) => void) {
  const socket = wx.connectSocket({
    url: wsURL + "/ws",
  });

  socket.onMessage((msg) => {
    onMsg(
      car.v1.CarEntity.fromObject(
        camelcaseKeys(JSON.parse(msg.data as string), { deep: true })
      )
    );
  });
  return socket;
}

export function getCar(id: string): Promise<car.v1.ICar> {
  return requestWithRetry({
    method: "GET",
    url: `/v1/car/${encodeURIComponent(id)}`,
    resolveRes: car.v1.Car.fromObject,
  });
}
