import $api from "../http";
import { IAuthResponse, ILoginRequest, IRegistrationRequest } from "../types/auth";

export default class AuthService {
  static async login(data: ILoginRequest): Promise<IAuthResponse> {
    const response = await $api.post<IAuthResponse>('/auth/login', data);
    return response.data;
  }

  static async registration(data: IRegistrationRequest): Promise<IAuthResponse> {
    const response = await $api.post<IAuthResponse>('/auth/register', data);
    return response.data;
  }

  static async logout(): Promise<void> {
    await $api.post('/auth/logout');
  }
}