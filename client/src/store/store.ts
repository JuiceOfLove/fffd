import { makeAutoObservable } from "mobx";
import AuthService from "../services/AuthService";
import { API_URL } from "../http";
import axios from "axios";
import { IUser, IAuthResponse } from "../types/auth";

export default class Store {
  user: IUser | null = null;
  isAuth: boolean = false;
  isLoading: boolean = false;
  authChecked: boolean = false; // новый флаг

  constructor() {
    makeAutoObservable(this);
  }

  setAuth(bool: boolean) {
    this.isAuth = bool;
  }

  setUser(user: IUser | null) {
    this.user = user;
  }

  setLoading(bool: boolean) {
    this.isLoading = bool;
  }

  setAuthChecked(bool: boolean) {
    this.authChecked = bool;
  }

  getCurrentUserId(): number | null {
    return this.user?.id || null;
  }

  async login(email: string, password: string) {
    try {
      const response: IAuthResponse = await AuthService.login({ email, password });
      localStorage.setItem('token', response.access_token);
      this.setAuth(true);
      this.setUser(response.user);
      this.setAuthChecked(true);
    } catch (e: any) {
      console.log(e.response?.data?.message);
      this.setAuthChecked(true);
    }
  }

  async registration(name: string, email: string, password: string) {
    try {
      const response: IAuthResponse = await AuthService.registration({ name, email, password });
      this.setAuth(true);
      this.setUser(response.user);
      this.setAuthChecked(true);
    } catch (e: any) {
      console.log(e.response?.data?.message);
      this.setAuthChecked(true);
    }
  }

  async logout() {
    try {
      await AuthService.logout();
      this.setAuth(false);
      this.setUser(null);
      this.setAuthChecked(true);
    } catch (e: any) {
      console.log(e.response?.data?.message);
    }
  }

  async checkAuth() {
    this.setLoading(true);
    try {
      const response = await axios.post<IAuthResponse>(
        `${API_URL}/auth/refresh`,
        {},
        { withCredentials: true }
      );
      localStorage.setItem("token", response.data.access_token);
      this.setAuth(true);
      this.setUser(response.data.user);
    } catch (e: any) {
      console.log("Ошибка checkAuth:", e.response?.data || e.message);
      this.setAuth(false); // если проверка не проходит, ставим неавторизованным
    } finally {
      this.setLoading(false);
      this.setAuthChecked(true); // завершаем проверку авторизации
    }
  }
}
