import { api, unwrap } from '@/lib/api'
import type { ApiResponse, ChannelModel } from '@/lib/types'

interface OpenAIListResponse<T> {
  object: 'list'
  data: T[]
}

export const modelsService = {
  /** Admin: full OpenAI-format model catalog across all channels. */
  async list(): Promise<ChannelModel[]> {
    // This endpoint does NOT use the standard envelope; it returns `{object, data}`.
    const res = await api.get<OpenAIListResponse<ChannelModel>>('/channel/models')
    return res.data?.data ?? []
  },

  /** Admin: channelType id → list of model names. */
  async byChannel(): Promise<Record<number, string[]>> {
    const res = await api.get<ApiResponse<Record<number, string[]>>>('/models')
    return unwrap(res) ?? {}
  },
}
