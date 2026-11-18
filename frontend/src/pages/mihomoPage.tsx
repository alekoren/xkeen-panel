const dashboard = 'http://192.168.2.1:9090/ui/'

export function MihomoPage() {
 return <section className='flex flex-col gap-3'>
  <div className='flex items-center justify-between gap-3'>
   <h1 className='text-xl font-semibold'>Mihomo</h1>
   <a href={dashboard} target='_blank' rel='noopener noreferrer' className='text-tint'>Открыть отдельно</a>
  </div>
  <p className='text-sm text-muted-foreground'>Встроена существующая панель на порту 9090. Если она не открывается, проверьте прямую ссылку. При активном Xray API Mihomo может быть недоступен. Вход в XKeen Panel не заменяет защиту API Mihomo.</p>
  {window.location.protocol === 'https:' ? <p role='alert'>Браузер может блокировать HTTP-панель внутри HTTPS. Используйте «Открыть отдельно» в домашней сети.</p> : <iframe src={dashboard} title='Панель Mihomo на порту 9090' referrerPolicy='no-referrer' className='w-full rounded-xl border border-white/15 bg-white' style={{ height: '78dvh', minHeight: 480 }} />}
 </section>
}
