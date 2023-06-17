import { rental } from "./proto_gen/rental/rental_pb";
import { requestWithRetry } from "../utils/index";

export function createTrip(
  req: rental.v1.ICreateTripRequest
): Promise<rental.v1.TripEntity> {
  return requestWithRetry({
    method: "POST",
    url: "/v1/trip",
    data: req,
    resolveRes: rental.v1.TripEntity.fromObject,
  });
}

// export async function getTrip(id: string): Promise<rental.v1.ITrip> {
//   return requestWithRetry({
//     method: "GET",
//     url: `/v1/trip/${encodeURIComponent(id)}`,
//     resolveRes: rental.v1.Trip.fromObject,
//   });
// }

// export function getTrips(
//   status?: rental.v1.TripStatus
// ): Promise<rental.v1.IGetTripsResponse> {
//   return requestWithRetry({
//     method: "GET",
//     url: `/v1/trips${status ? "?status=" + status : ""}`,
//     resolveRes: rental.v1.GetTripsResponse.fromObject,
//   });
// }

export async function getTrip(id: string): Promise<rental.v1.ITrip> {
  const { trips } = await getTrips([id]);
  return trips !== undefined && trips !== null
    ? Promise.resolve(trips[0].trip!)
    : Promise.reject(null);
}

export function getTrips(
  idList: string[] = []
): Promise<rental.v1.IGetTripsResponse> {
  return requestWithRetry({
    method: "POST",
    url: "/v1/trips",
    data: {
      idList,
    },
    resolveRes: rental.v1.GetTripsResponse.fromObject,
  });
}

export function updateTripPos(id: string, loc?: rental.v1.ILocation) {
  return updateTrip({
    id,
    current: loc,
  });
}

export function finishTrip(id: string) {
  return updateTrip({
    id,
    endTrip: true,
  });
}

function updateTrip(r: rental.v1.IUpdateTripRequest): Promise<rental.v1.ITrip> {
  if (!r.id) {
    return Promise.reject("must specify id");
  }
  return requestWithRetry({
    method: "PUT",
    url: `/v1/trip/${encodeURIComponent(r.id)}`,
    data: r,
    resolveRes: rental.v1.Trip.fromObject,
  });
}
