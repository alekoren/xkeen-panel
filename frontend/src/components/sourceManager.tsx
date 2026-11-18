import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'

type Source = { id: string; name: string; kind: string; status: string }
export function SourceManager() {
 const qc = useQueryClient()
 const [name, setName] = useState('')
 const [kind, setKind] = useState('subscription')
 const [payload, setPayload] = useState('')
 const [fileError, setFileError] = useState('')
 const sources = useQuery({ queryKey: ['sources'], queryFn: () => api.get<Source[]>('/api/sources') })
 const create = useMutation({
  mutationFn: () => api.post('/api/sources', { name, kind, payload }),
  onSuccess: () => { setName(''); setPayload(''); void qc.invalidateQueries({ queryKey: ['sources'] }) },
 })
 return <section className='rounded-xl border border-white/15 p-4 flex flex-col gap-3'>
  <h2>Источники — каталог импорта</h2>
  <p className='text-sm text-muted-foreground'>Сохранение не включает VPN и не меняет конфигурацию. Загрузка подписок и применение источника ещё недоступны. Текущие провайдеры Mihomo не импортируются автоматически.</p>
  {sources.isPending && <p>Загрузка списка…</p>}
  {(sources.data ?? []).map(source => <div key={source.id} className='border-b border-white/10 py-2'>
   <div>{source.name}</div><div className='text-sm text-muted-foreground'>{source.kind === 'subscription' ? 'Ссылка подписки' : 'YAML Mihomo'} · Сохранён, не применён</div>
  </div>)}
  {sources.data?.length === 0 && <p>Сохранённых источников пока нет.</p>}
  <form className='flex flex-col gap-3' onSubmit={event => { event.preventDefault(); create.mutate() }}>
   <label>Название<input required maxLength={100} className='block w-full rounded border p-2 bg-transparent' value={name} onChange={event => setName(event.target.value)} /></label>
   <label>Тип<select className='block w-full rounded border p-2 bg-transparent' value={kind} onChange={event => { setKind(event.target.value); setPayload(''); setFileError(''); create.reset() }}>
    <option value='subscription'>HTTPS-ссылка подписки (только сохранение)</option><option value='mihomo-yaml'>Файл YAML Mihomo с proxies</option>
   </select></label>
   {kind === 'subscription' ? <label>Ссылка<input required type='url' autoComplete='off' spellCheck={false} className='block w-full rounded border p-2 bg-transparent' value={payload} onChange={event => setPayload(event.target.value)} /></label> : <label>Файл .yaml / .yml<input type='file' accept='.yaml,.yml' disabled={create.isPending} onChange={async event => {
    const file = event.target.files?.[0]; setPayload(''); setFileError('')
    if (!file) return
    if (file.size > 1048576) { setFileError('Максимум 1 МиБ'); return }
    try { setPayload(await file.text()) } catch { setFileError('Не удалось прочитать файл') }
   }} /></label>}
   <button type='submit' className='rounded border p-2 disabled:opacity-50' disabled={create.isPending || !name.trim() || !payload}>{create.isPending ? 'Сохраняю…' : 'Сохранить без применения'}</button>
  </form>
  {(create.error || sources.error || fileError) && <p role='alert'>{fileError || create.error?.message || sources.error?.message}</p>}
  {create.isSuccess && <p role='status'>Источник сохранён. Работающий VPN не изменён.</p>}
  <p className='text-sm text-muted-foreground'>Импорт WireGuard/Proton .conf пока не реализован. Проверка структуры YAML не заменяет проверку ядром.</p>
 </section>
}
