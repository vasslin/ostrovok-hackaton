export type MoneyMinor = number;
export type LocalDate = string;
export type ZonedDateTime = string;
export interface Criteria {
  origin: string;
  destination: string;
  startDate: LocalDate;
  nights: number;
  adults: number;
  childrenAges: number[];
  budgetMinor: MoneyMinor;
  currency: 'RUB';
  pets: boolean;
}
export interface Ticket {
  id: string;
  carrier: string;
  number: string;
  type: 'flight';
  departure: ZonedDateTime;
  arrival: ZonedDateTime;
  priceMinor: MoneyMinor;
  baggage: boolean;
  refundable: boolean;
  petsAllowed: boolean;
  demo: boolean;
}
export interface Hotel {
  id: string;
  name: string;
  city: string;
  rating: number;
  stars: number;
  priceMinor: MoneyMinor;
  nightlyMinor: MoneyMinor;
  nights: number;
  rooms: number;
  petsAllowed: boolean;
  cancellation: string;
  meals: string;
  distance: string;
  cover: 'sea' | 'city' | 'mountain';
  demo: boolean;
}
export interface Activity { id: string; name: string; duration: string; priceMinor: MoneyMinor; note: string; demo: boolean; }
export interface Stop {
  id: string;
  city: string;
  country: string;
  lat: number;
  lng: number;
  arrival: ZonedDateTime;
  departure: ZonedDateTime;
  nights: number;
  hotelOptions: Hotel[];
  hotel: Hotel | null;
  activities: Activity[];
}
export interface Segment {
  id: string;
  fromStopId: string;
  toStopId: string;
  from: string;
  to: string;
  options: Ticket[];
  ticket: Ticket;
}
export interface Trip {
  id: string;
  revision: number;
  status: 'draft' | 'booked';
  demo: boolean;
  title: string;
  subtitle: string;
  criteria: Criteria;
  startDate: LocalDate;
  endDate: LocalDate;
  stops: Stop[];
  segments: Segment[];
  totals: { currency: 'RUB'; transportMinor: MoneyMinor; hotelsMinor: MoneyMinor; transferMinor: MoneyMinor; activitiesMinor: MoneyMinor; totalMinor: MoneyMinor };
  warnings: string[];
  tags: string[];
  cover: 'sea' | 'city' | 'mountain';
}
export interface ChatResponse { trip: Trip | null; message: string; actions: { label: string; value: string }[]; }
export interface Profile { name: string; email: string; phone: string; }
export interface Booking { id: string; status: 'confirmed_demo'; createdAt: ZonedDateTime; demo: true; }
export interface HistoryItem { booking: Booking; trip: Trip; }
