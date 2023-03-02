import { rental } from "./proto_gen/rental/rental_pb";
import { requestWithRetry } from "../utils/index";

export function createTrip(
  req: rental.v1.ICreateTripRequest
): Promise<rental.v1.ICreateTripResponse> {
  return requestWithRetry({
    method: "POST",
    url: "/v1/trip",
    data: req,
    resolveRes: rental.v1.CreateTripResponse.fromObject,
  });
}

// export function getTrip(id: string): Promise<rental.v1.ITrip> {
//   return requestWithRetry({
//     method: "GET",
//     url: `/v1/trip/${encodeURIComponent(id)}`,
//     resolveRes: rental.v1.Trip.fromObject,
//   });
// }

// export function getTrips(
//   s?: rental.v1.TripStatus
// ): Promise<rental.v1.IGetTripsResponse> {
//   let url = "/v1/trips";
//   if (s) {
//     url += `?status=${s}`;
//   }
//   return requestWithRetry({
//     method: "GET",
//     url,
//     resolveRes: rental.v1.GetTripsResponse.fromObject,
//   });
// }

// export function updateTripPos(id: string, loc?: rental.v1.ILocation) {
//   return updateTrip({
//     id,
//     current: loc,
//   });
// }

// export function finishTrip(id: string) {
//   return updateTrip({
//     id,
//     endTrip: true,
//   });
// }

// function updateTrip(
//   r: rental.v1.IUpdateTripRequest
// ): Promise<rental.v1.ITrip> {
//   if (!r.id) {
//     return Promise.reject("must specify id");
//   }
//   return requestWithRetry({
//     method: "PUT",
//     url: `/v1/trip/${encodeURIComponent(r.id)}`,
//     data: r,
//     resolveRes: rental.v1.Trip.fromObject,
//   });
// }
