"use client";

import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";

type Language = "ru" | "en";
const STORAGE_KEY = "openinvest.ui.language";

const translations: Record<string, string> = {
  "Dashboard": "Обзор",
  "Portfolios": "Портфели",
  "Assets": "Активы",
  "Dividend calculator": "Калькулятор дивидендов",
  "Corporate actions": "Корпоративные события",
  "OpenInvest workspace": "Рабочее пространство OpenInvest",
  "Personal capital": "Личный капитал",
  "Refresh": "Обновить",
  "Refreshing…": "Обновление…",
  "Sign out": "Выйти",
  "Signing out…": "Выход…",
  "Skip to main content": "Перейти к содержимому",
  "Primary navigation": "Основная навигация",
  "Mobile navigation": "Мобильная навигация",
  "OpenInvest account": "Учётная запись OpenInvest",
  "Sign in to your private capital workspace.": "Управляйте своим капиталом в OpenInvest.",
  "Sign in or create an account to access your private portfolios. Your session is protected with secure cookies.": "Войдите или зарегистрируйтесь, чтобы управлять личными портфелями.",
  "Authentication mode": "Режим авторизации",
  "Login": "Войти",
  "Register": "Регистрация",
  "Email": "Электронная почта",
  "Password": "Пароль",
  "Sending...": "Отправка…",
  "Create account": "Создать аккаунт",
  "If this address is eligible, instructions will be sent.": "Если адрес подходит для регистрации, инструкции будут отправлены на почту.",
  "Personal Capital Operating System": "Управление личным капиталом",
  "Track portfolios, returns and dividends in one place.": "Портфели, доходность и дивиденды — в одном месте.",
  "Review your portfolios, investment activity, income and returns with OpenInvest.": "Следите за портфелями, инвестиционными операциями, доходами и доходностью.",
  "Loading portfolios…": "Загрузка портфелей…",
  "Portfolio data unavailable": "Данные портфелей недоступны",
  "We could not load your portfolios right now. Try again in a moment.": "Портфели пока не загрузились. Повторите попытку позже.",
  "Portfolio list": "Список портфелей",
  "Your portfolios": "Ваши портфели",
  "Discover assets": "Найти активы",
  "Reload": "Перезагрузить",
  "Load more portfolios": "Показать ещё портфели",
  "Supported instruments": "Доступные инструменты",
  "Asset discovery": "Поиск активов",
  "Search supported assets before adding transactions.": "Найдите доступные активы перед добавлением операций.",
  "First portfolio": "Первый портфель",
  "Create a RUB portfolio": "Создать портфель в рублях",
  "Create a RUB portfolio to record transactions, track positions and review performance.": "Создайте портфель для учёта операций, отслеживания позиций и доходности.",
  "Portfolio name": "Название портфеля",
  "Create portfolio": "Создать портфель",
  "Creating…": "Создание…",
  "Find supported stocks and bonds.": "Найдите акции и облигации.",
  "Search supported assets by ticker or name. Missing market data is left unavailable rather than estimated.": "Ищите инструменты по тикеру или названию. Недоступные рыночные данные не заменяются оценками.",
  "Back to portfolios": "К портфелям",
  "Search by ticker or name": "Тикер или название",
  "Asset type": "Тип актива",
  "All supported types": "Все типы",
  "Stocks": "Акции",
  "Bonds": "Облигации",
  "Search": "Найти",
  "Searching...": "Поиск…",
  "Loading supported assets...": "Загрузка активов…",
  "No supported assets matched this search.": "По запросу ничего не найдено.",
  "Last price unavailable": "Цена недоступна",
  "Load more": "Показать ещё",
  "Loading more...": "Загрузка…",
  "Close": "Закрыть",
  "Calculate gross dividend income from your inputs.": "Рассчитайте дивидендный доход.",
  "Enter your position and dividend inputs. OpenInvest calculates gross dividend income and, when position cost is provided, gross yield. Tax is not included.": "Введите параметры позиции и дивидендов. OpenInvest рассчитает доход до налогообложения и доходность при наличии стоимости позиции.",
  "Ticker": "Тикер",
  "Quantity": "Количество",
  "Dividend per share, ₽": "Дивиденд на акцию, ₽",
  "Position cost, ₽": "Стоимость позиции, ₽",
  "(optional)": "(необязательно)",
  "Calculate": "Рассчитать",
  "Calculating…": "Расчёт…",
  "Result": "Результат",
  "Enter values and calculate.": "Введите значения и выполните расчёт.",
  "Calculation rejected": "Расчёт не выполнен",
  "Dividend per share": "Дивиденд на акцию",
  "Gross dividend": "Дивиденды до вычета налогов",
  "Position cost": "Стоимость позиции",
  "Gross dividend yield": "Валовая дивидендная доходность",
  "Methodology": "Методика",
  "Gross · tax not included": "До вычета налогов",
  "Tax included": "С учётом налогов",
  "Dividend and coupon calendar with an event activity heatmap.": "Календарь дивидендов и купонов.",
  "If corporate-action data is unavailable, OpenInvest says so instead of showing an empty calendar. Dates and amounts are shown only when supplied by the data source.": "Календарь показывает только данные, полученные от источника. При отсутствии данных отображается уведомление.",
  "Instruments": "Инструменты",
  "From": "С",
  "To": "По",
  "Enter one or more instrument identifiers, separated by commas or spaces.": "Укажите тикеры через запятую или пробел.",
  "Load calendar": "Загрузить календарь",
  "Loading...": "Загрузка…",
  "Choose an instrument set and date window.": "Выберите инструменты и период.",
  "Loading corporate actions...": "Загрузка событий…",
  "Corporate actions source unavailable.": "Источник данных недоступен.",
  "Corporate-action data is temporarily unavailable. No events are shown until a data source responds.": "Данные о корпоративных событиях временно недоступны.",
  "No dated events found for this selection.": "На выбранные даты событий не найдено.",
  "The data source responded successfully and returned no events.": "Источник ответил, но событий за период нет.",
  "Calendar": "Календарь",
  "Current dated events": "События по датам",
  "Heatmap": "Тепловая карта",
  "Event density": "Частота событий",
  "Portfolio detail": "Портфель",
  "Loading portfolio…": "Загрузка портфеля…",
  "Transaction history": "История операций",
  "Transactions": "Операции",
  "No transactions yet. Add the first transaction above.": "Операций пока нет. Добавьте первую операцию выше.",
  "Portfolio on": "Портфель на",
  "Positions": "Позиции",
  "Total acquisition basis": "Общая стоимость приобретения",
  "Current": "Текущие",
  "Historical": "История",
  "Historical date": "Дата в прошлом",
  "Portfolio time machine": "История портфеля",
  "Cost basis and manual valuation": "Стоимость приобретения и ручная оценка",
  "Viewing current positions. Manual prices are user-supplied inputs, not live quotes.": "Показаны текущие позиции. Цены введены пользователем и не являются биржевыми котировками.",
  "Choose a historical date to reconstruct your positions.": "Выберите дату для просмотра позиций в прошлом.",
  "Manual valuation coverage": "Полнота ручной оценки",
  "Positions unavailable": "Позиции недоступны",
  "Unavailable values are left unavailable rather than estimated.": "Недоступные значения не заменяются приблизительными оценками.",
  "No open stock or bond positions.": "Открытых позиций по акциям или облигациям нет.",
  "Asset": "Актив",
  "Average acquisition price": "Средняя цена покупки",
  "Acquisition basis": "Стоимость приобретения",
  "Acquisition-basis weight": "Доля стоимости приобретения",
  "Market valuation": "Рыночная оценка",
  "Market valuation unavailable": "Рыночная оценка недоступна",
  "Portfolio unavailable": "Портфель недоступен",
  "Historical positions only": "Только исторические позиции",
  "Summary not available": "Сводка недоступна",
  "Type": "Тип",
  "Trade date": "Дата операции",
  "Amount": "Сумма",
  "Status": "Статус",
  "Actions": "Действия",
  "Loading transactions…": "Загрузка операций…",
  "Load more transactions": "Показать ещё операции",
  "Portfolio capital": "Капитал портфеля",
  "Capital summary": "Сводка капитала",
  "Total capital (cost basis)": "Общий капитал по стоимости приобретения",
  "Cash": "Денежные средства",
  "Stocks (cost basis)": "Акции по стоимости приобретения",
  "Bonds (cost basis)": "Облигации по стоимости приобретения",
  "Invested capital": "Инвестированный капитал",
  "Gross dividends recorded": "Начисленные дивиденды до вычета налогов",
  "Gross coupons recorded": "Начисленные купоны до вычета налогов",
  "Portfolio activity": "Операции портфеля",
  "Add transaction": "Добавить операцию",
  "Transaction type": "Тип операции",
  "Unit price": "Цена за единицу",
  "Gross amount": "Сумма до удержаний",
  "Commission": "Комиссия",
  "Tax": "Налог",
  "Settlement date": "Дата расчёта",
  "Note": "Примечание",
  "Saving…": "Сохранение…",
  "BUY": "Покупка",
  "SELL": "Продажа",
  "DIVIDEND": "Дивиденды",
  "COUPON": "Купоны",
  "FEE": "Комиссия",
  "TAX": "Налог",
  "DEPOSIT": "Пополнение",
  "WITHDRAWAL": "Вывод средств",
};

type LocaleContextValue = {
  language: Language;
  setLanguage: (value: Language) => void;
  t: (english: string) => string;
};
const LocaleContext = createContext<LocaleContextValue | null>(null);
const isolatedComponentFallback: LocaleContextValue = { language: "en", setLanguage: () => {}, t: (english) => english };

export function LocaleProvider({ children }: { children: ReactNode }) {
  const [language, setLanguage] = useState<Language>("ru");
  useEffect(() => {
    try {
      const saved = window.localStorage.getItem(STORAGE_KEY);
      if (saved === "ru" || saved === "en") setLanguage(saved);
    } catch { /* Storage can be blocked by browser settings. */ }
  }, []);
  useEffect(() => {
    document.documentElement.lang = language;
    try { window.localStorage.setItem(STORAGE_KEY, language); } catch { /* Browsers may block storage. */ }
  }, [language]);
  const value = useMemo<LocaleContextValue>(() => ({
    language,
    setLanguage,
    t: (english) => language === "ru" ? translations[english] ?? english : english,
  }), [language]);
  return <LocaleContext.Provider value={value}>{children}</LocaleContext.Provider>;
}

export function useLocale(): LocaleContextValue {
  const value = useContext(LocaleContext);
  return value ?? isolatedComponentFallback;
}

export function LanguageSwitch() {
  const { language, setLanguage } = useLocale();
  return (
    <div role="group" aria-label={language === "ru" ? "Язык интерфейса" : "Interface language"} className="oi-language-switch">
      <button type="button" onClick={() => setLanguage("ru")} aria-pressed={language === "ru"} lang="ru">RU</button>
      <button type="button" onClick={() => setLanguage("en")} aria-pressed={language === "en"} lang="en">EN</button>
    </div>
  );
}
