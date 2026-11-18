import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Group, Row } from '@/components/ios'
import { SourceManager } from '@/components/sourceManager'
import { api } from '@/lib/api'

type Core = 'xray' | 'mihomo'
type CoreState = { core: string; mode: string; available: boolean; running: boolean; busy: boolean; target: string; error: string }
const key = ['xkeen-core'] as const
const names: Record<Core, string> = { xray: 'Xray', mihomo: 'Mihomo' }

export function CoreSwitch() {
 const qc = useQueryClient()
 const state = useQuery({ queryKey: key, queryFn: () => api.get<CoreState>('/api/xkeen/core'), refetchInterval: query => query.state.data?.busy ? 1500 : 8000 })
 const change = useMutation({
  mutationFn: (input: { target: Core; expected: Core }) => api.post('/api/xkeen/core', input),
  onSuccess: (_, input) => qc.setQueryData<CoreState>(key, old => old ? { ...old, busy: true, target: input.target, error: '' } : old),
  onSettled: () => qc.invalidateQueries({ queryKey: key }),
 })
 const s = state.data
 const current = s?.core === 'xray' || s?.core === 'mihomo' ? s.core : undefined
 const target = current === 'xray' ? 'mihomo' : 'xray'
 const error = change.error?.message ?? state.error?.message ?? s?.error
 const detail = s?.busy ? 'Выполняется переключение или перезапуск XKeen…' : !s ? 'Читаю состояние XKeen…' : !s.available ? 'Переключение доступно в XKeen 2.x.' : !s.running ? 'Сначала запустите текущее ядро.' : undefined
 return <>
  <Group title='Ядро прокси' footer={error || detail ? <span role={error ? 'alert' : 'status'}>{error || detail}</span> : undefined}>
   <Row label={current ? names[current] : 'Ядро не определено'} detail={s?.mode} />
   <Row label={change.isPending || s?.busy ? 'Подождите…' : `Переключить на ${names[target]}`} tone='tint'
    disabled={!current || !s?.available || !s.running || s.busy || state.isError || change.isPending}
    onClick={() => { if (current && window.confirm(`Переключить на ${names[target]}? Текущие прокси-соединения могут прерваться.`)) change.mutate({ target, expected: current }) }} />
  </Group>
  <SourceManager />
 </>
}
