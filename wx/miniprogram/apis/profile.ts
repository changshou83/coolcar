import { rental } from "./proto_gen/rental/rental_pb";
import { requestWithRetry } from "../utils/index";

// Profile
export function getProfile(): Promise<rental.v1.IProfile> {
  return requestWithRetry({
    method: "GET",
    url: "/v1/profile",
    resolveRes: rental.v1.Profile.fromObject,
  });
}

export function submitProfile(
  identity: rental.v1.IIdentity
): Promise<rental.v1.IProfile> {
  return requestWithRetry({
    method: "POST",
    url: "/v1/profile",
    data: identity,
    resolveRes: rental.v1.Profile.fromObject,
  });
}

export function clearProfile(): Promise<rental.v1.IProfile> {
  return requestWithRetry({
    method: "DELETE",
    url: `/v1/profile`,
    resolveRes: rental.v1.Profile.fromObject,
  });
}
