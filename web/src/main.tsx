import React from 'react'
import ReactDOM from 'react-dom/client'
import Home from '@/pages/home.tsx'
import './index.css'
import {BrowserRouter, Route, Routes} from "react-router-dom";
import {Booking} from "@/pages/booking.tsx";
import {Schedule} from "@/pages/schedule.tsx";

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
      <BrowserRouter basename="fe">
          <Routes>
              <Route path="/" element={<Home/>} />
              <Route path="/booking/:username" element={<Booking/>} />
              <Route path="/schedule/:id" element={<Schedule/>} />
              <Route path="*" element={<>
                  <div className="flex flex-col my-12 text-center">
                      <h1 className="text-lg font-bold mb-2">Not Found</h1>
                      <span className="text-sm">
                          do you mean?
                          <a href="/fe/booking/mentor">
                              <code className="ml-1 underline italic text-blue-400">
                                  [url]/booking/[username]
                              </code>
                          </a>
                      </span>
                  </div>
              </>}/>
          </Routes>
      </BrowserRouter>
  </React.StrictMode>,
)
