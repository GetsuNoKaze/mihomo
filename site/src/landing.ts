// Texts of the landing page, the header and the footer in both languages. Docs pages live
// in src/content/docs (Russian) and src/content/docs/en (English); news posts in
// src/content/news/<lang>.

export const links = {
	github: 'https://github.com/getmikan/Citrus',
	releases: 'https://github.com/getmikan/Citrus/releases',
	latest: 'https://github.com/getmikan/Citrus/releases/latest',
	prereleaseMain: 'https://github.com/getmikan/Citrus/releases/tag/Prerelease-main',
	prereleaseDev: 'https://github.com/getmikan/Citrus/releases/tag/Prerelease-dev',
	issues: 'https://github.com/getmikan/Citrus/issues',
	license: 'https://github.com/getmikan/Citrus/blob/main/LICENSE',
	config: 'https://github.com/getmikan/Citrus/blob/main/docs/config.yaml',
	mikan: 'https://miroshka000.github.io/mikan/',
	mikanRepo: 'https://github.com/Miroshka000/mikan',
	mihomo: 'https://github.com/MetaCubeX/mihomo',
};

export const runCommand = 'citrus -d ~/.config/citrus';

export type Lang = 'ru' | 'en';

const base = import.meta.env.BASE_URL.replace(/\/$/, '');
/** A page of this site in the given language: page('en', 'news/') → /Citrus/en/news/. */
export const page = (lang: Lang, path = '') => `${base}/${lang === 'en' ? 'en/' : ''}${path}`;

export const dateFormat = (lang: Lang, date: Date) =>
	date.toLocaleDateString(lang === 'en' ? 'en-GB' : 'ru-RU', { day: 'numeric', month: 'long', year: 'numeric', timeZone: 'UTC' });

export const strings = {
	ru: {
		hero: {
			title: 'Сетевое ядро Mikan',
			lead: 'Citrus — кроссплатформенное прокси-ядро: листенеры, TUN, шифрованный DNS и маршрутизация по правилам в одном YAML-конфиге. Форк mihomo, который развивается вместе с панелью Mikan. GPL-3.0.',
			note: 'Один бинарник для Linux, Windows, macOS, FreeBSD и Android.',
			primary: 'Начать',
			secondary: 'Скачать',
			art: 'Лайм, символ Citrus',
		},
		features: {
			title: 'Что умеет',
			items: [
				{
					title: 'Входящие',
					text: 'HTTP, HTTPS, SOCKS, mixed, redirect и TPROXY. Плюс серверные листенеры для VLESS, Trojan, Hysteria2, TUIC, AnyTLS и других.',
				},
				{
					title: 'Исходящие протоколы',
					text: 'VLESS с REALITY, Trojan, Shadowsocks, Hysteria2, TUIC, AnyTLS, TrustTunnel, Snell, WireGuard и другие. VMess пока есть, но его планируем убрать.',
				},
				{
					title: 'TUN',
					text: 'Весь трафик системы через ядро, с настраиваемым сетевым стеком.',
				},
				{
					title: 'DNS',
					text: 'DNS over HTTPS, TLS и QUIC, fake IP и правила маршрутизации для DNS-запросов.',
				},
				{
					title: 'Правила',
					text: 'Маршрутизация по домену, диапазону IP, geodata, процессу и логическим комбинациям правил.',
				},
				{
					title: 'Провайдеры и группы',
					text: 'Удалённые провайдеры прокси и правил, проверки доступности, fallback и балансировка нагрузки.',
				},
				{
					title: 'API контроллера',
					text: 'HTTP и WebSocket API, совместимые с клиентами и дашбордами для mihomo, например MetaCubeXD.',
				},
			],
		},
		download: {
			title: 'Скачать',
			lead: 'Готовые сборки лежат в релизах на GitHub. Файлы называются',
			naming: 'citrus-<os>-<arch>-<версия>',
			naming2: ': в релизе версия — это тег (citrus-linux-arm64-v1.2.3.gz), в пререлизе — ветка и коммит (citrus-linux-arm64-dev-1a2b3c4.gz). Для Windows это .zip, для Linux есть ещё .deb, .rpm и .pkg.tar.zst.',
			headers: ['Канал', 'Откуда', 'Что это'],
			channels: [
				['release', 'Последний тег vX.Y.Z', 'Стабильные версии.'],
				['Prerelease-main', 'Ветка main', 'Сборка с каждого слияния в main.'],
				['Prerelease-dev', 'Ветка dev', 'Самые свежие изменения, сюда приходят все PR.'],
			],
			arch: 'Файл amd64 собран под x86-64-v3 (Intel Haswell, AMD Excavator и новее); для старых берите amd64-compatible (v1) или amd64-v2. Есть также arm64, armv5–v7, 386, mips, riscv64, loong64, s390x и другие.',
			button: 'Все релизы',
			docs: 'Как запустить',
		},
		mikan: {
			title: 'Citrus и Mikan',
			text: 'Mikan — свободная VPN-панель. Её ноды запускают Citrus как библиотеку и считают трафик по каждому соединению каждого пользователя. Ядро живёт в отдельном репозитории, чтобы его можно было использовать и без панели: как обычный клиент или сервер в стиле mihomo.',
			site: 'Сайт Mikan',
			code: 'Mikan на GitHub',
		},
		news: {
			title: 'Новости',
			all: 'Все новости',
			rss: 'RSS',
		},
		header: {
			label: 'Проект',
			githubLabel: 'GitHub: исходный код Citrus',
			news: 'Новости',
			download: 'Скачать',
			downloadLabel: 'Скачать Citrus: релизы на GitHub',
			otherLanguage: 'English version',
		},
		footer: {
			label: 'Ссылки проекта',
			license: 'Citrus: свободная программа под GNU GPL v3, форк MetaCubeX/mihomo.',
			releases: 'Релизы',
			issues: 'Сообщить о проблеме',
			mikan: 'Mikan',
		},
		newsPage: {
			title: 'Новости',
			description: 'Новости Citrus: релизы, изменения в протоколах и планы.',
			empty: 'Пока новостей нет.',
			back: 'Все новости',
			feed: 'Подписаться через RSS',
		},
	},
	en: {
		hero: {
			title: 'The network core of Mikan',
			lead: 'Citrus is a cross-platform proxy core: listeners, TUN, encrypted DNS and rule-based routing in one YAML configuration. A fork of mihomo developed together with the Mikan panel. GPL-3.0.',
			note: 'One binary for Linux, Windows, macOS, FreeBSD and Android.',
			primary: 'Get started',
			secondary: 'Download',
			art: 'A lime, the Citrus mascot',
		},
		features: {
			title: 'What it does',
			items: [
				{
					title: 'Inbound',
					text: 'HTTP, HTTPS, SOCKS, mixed, redirect and TPROXY. Plus server listeners for VLESS, Trojan, Hysteria2, TUIC, AnyTLS and more.',
				},
				{
					title: 'Outbound protocols',
					text: 'VLESS with REALITY, Trojan, Shadowsocks, Hysteria2, TUIC, AnyTLS, TrustTunnel, Snell, WireGuard and more. VMess is still there, but we plan to remove it.',
				},
				{
					title: 'TUN',
					text: 'All system traffic through the core, with a configurable network stack.',
				},
				{
					title: 'DNS',
					text: 'DNS over HTTPS, TLS and QUIC, fake IP and routing rules for DNS queries.',
				},
				{
					title: 'Rules',
					text: 'Routing by domain, IP range, geodata, process and logical combinations of rules.',
				},
				{
					title: 'Providers and groups',
					text: 'Remote proxy and rule providers, health checks, fallback and load balancing.',
				},
				{
					title: 'Controller API',
					text: 'HTTP and WebSocket APIs compatible with mihomo clients and dashboards such as MetaCubeXD.',
				},
			],
		},
		download: {
			title: 'Download',
			lead: 'Builds are published as GitHub releases. Files are named',
			naming: 'citrus-<os>-<arch>-<version>',
			naming2: ': in a release the version is the tag (citrus-linux-arm64-v1.2.3.gz), in a prerelease the branch and commit (citrus-linux-arm64-dev-1a2b3c4.gz). Windows builds are .zip; Linux also gets .deb, .rpm and .pkg.tar.zst.',
			headers: ['Channel', 'Built from', 'What it is'],
			channels: [
				['release', 'The latest vX.Y.Z tag', 'Stable versions.'],
				['Prerelease-main', 'The main branch', 'A build of every merge into main.'],
				['Prerelease-dev', 'The dev branch', 'The newest changes; every PR lands here.'],
			],
			arch: 'The amd64 file targets x86-64-v3 (Intel Haswell, AMD Excavator and newer); for older ones take amd64-compatible (v1) or amd64-v2. There are also arm64, armv5–v7, 386, mips, riscv64, loong64, s390x and more.',
			button: 'All releases',
			docs: 'How to run it',
		},
		mikan: {
			title: 'Citrus and Mikan',
			text: 'Mikan is a free VPN panel. Its nodes run Citrus as a library and count traffic per user on every connection. The core lives in its own repository so it can be used without the panel too: as a regular mihomo-style client or server.',
			site: 'Mikan website',
			code: 'Mikan on GitHub',
		},
		news: {
			title: 'News',
			all: 'All news',
			rss: 'RSS',
		},
		header: {
			label: 'Project',
			githubLabel: 'GitHub: the Citrus source code',
			news: 'News',
			download: 'Download',
			downloadLabel: 'Download Citrus: releases on GitHub',
			otherLanguage: 'Русская версия',
		},
		footer: {
			label: 'Project links',
			license: 'Citrus is free software under the GNU GPL v3, a fork of MetaCubeX/mihomo.',
			releases: 'Releases',
			issues: 'Report a problem',
			mikan: 'Mikan',
		},
		newsPage: {
			title: 'News',
			description: 'Citrus news: releases, protocol changes and plans.',
			empty: 'No news yet.',
			back: 'All news',
			feed: 'Subscribe via RSS',
		},
	},
} as const;
