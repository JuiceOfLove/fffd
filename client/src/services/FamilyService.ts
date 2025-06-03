import $api from "../http";
import { IFamilyResponse, IInviteResponse } from "../types/family";

export default class FamilyService {
  // Создание семьи
  static async createFamily(name: string): Promise<IFamilyResponse> {
    const response = await $api.post<IFamilyResponse>('/family/create', { name });
    return response.data;
  }

  // Отправка приглашения
  static async inviteMember(email: string): Promise<IInviteResponse> {
    const response = await $api.post<IInviteResponse>('/family/invite', { email });
    return response.data;
  }

  // Принятие приглашения
  static async acceptInvitation(token: string): Promise<IInviteResponse> {
    const response = await $api.get<IInviteResponse>(`/family/accept/${token}`);
    return response.data;
  }

  // Получение деталей семьи: объект семьи и список её членов
  static async getFamilyDetails() {
    const response = await $api.get<IFamilyResponse>("/family/details");
    return response.data;
  }
}