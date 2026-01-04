package handlers

import (
	"fmt"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="uk">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Безкоштовна консультація з Христина Івасюк</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        :root {
            /* Burgundy Theme */
            --primary-start: #800020;
            --primary-end: #5c0011;
            --accent-color: #a0153e;
            --gradient-bg: linear-gradient(135deg, #800020 0%, #5c0011 100%);
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: var(--gradient-bg);
            min-height: 100vh;
            padding: 20px;
            transition: background 0.3s ease;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }

        .header {
            background: var(--gradient-bg);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 2.5rem;
            margin-bottom: 10px;
        }

        .header p {
            font-size: 1.1rem;
            opacity: 0.9;
        }

        .content {
            padding: 40px;
        }

        .calendar-grid {
            display: grid;
            grid-template-columns: repeat(7, 1fr);
            gap: 10px;
            margin-bottom: 30px;
        }

        .day-cell {
            padding: 20px 10px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s ease;
            background: white;
            min-height: 80px;
            display: flex;
            flex-direction: column;
            justify-content: center;
        }

        .day-cell:hover {
            border-color: var(--primary-start);
            background: #f0f4ff;
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
        }

        .day-cell.selected {
            border-color: var(--primary-start);
            background: var(--primary-start);
            color: white;
        }

        .day-cell.no-slots {
            background: #f5f5f5;
            color: #999;
            cursor: not-allowed;
            opacity: 0.5;
        }

        .day-cell.no-slots:hover {
            transform: none;
            box-shadow: none;
            background: #f5f5f5;
            border-color: #e0e0e0;
        }

        .day-number {
            font-size: 1.5rem;
            font-weight: bold;
            margin-bottom: 5px;
        }

        .day-name {
            font-size: 0.8rem;
            opacity: 0.8;
            text-transform: uppercase;
        }

        .day-slots-count {
            font-size: 0.75rem;
            margin-top: 5px;
            opacity: 0.9;
        }

        .time-slots-panel {
            background: #f9f9f9;
            border-radius: 12px;
            padding: 30px;
            margin-bottom: 30px;
            display: none;
        }

        .time-slots-panel.active {
            display: block;
        }

        .time-slots-header {
            text-align: center;
            margin-bottom: 20px;
        }

        .time-slots-header h3 {
            color: #333;
            font-size: 1.3rem;
            margin-bottom: 5px;
        }

        .time-slots-header p {
            color: #666;
            font-size: 0.9rem;
        }

        .time-slots-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
            gap: 10px;
            margin-bottom: 20px;
        }

        .time-slot {
            padding: 15px 10px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s ease;
            background: white;
            font-size: 1rem;
            font-weight: 600;
        }

        .time-slot:hover {
            border-color: var(--primary-start);
            background: #f0f4ff;
            transform: translateY(-2px);
        }

        .time-slot.booked {
            background: #f5f5f5;
            color: #999;
            cursor: not-allowed;
            opacity: 0.6;
            text-decoration: line-through;
        }

        .time-slot.booked:hover {
            transform: none;
            border-color: #e0e0e0;
            background: #f5f5f5;
        }

        .back-btn {
            padding: 10px 20px;
            background: #e0e0e0;
            color: #666;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-weight: 600;
            transition: all 0.3s;
            display: block;
            margin: 0 auto 20px;
        }

        .back-btn:hover {
            background: #d0d0d0;
        }

        .booking-form {
            max-width: 500px;
            margin: 30px auto;
            padding: 30px;
            background: #f9f9f9;
            border-radius: 12px;
            display: none;
        }

        .booking-form.active {
            display: block;
        }

        .form-group {
            margin-bottom: 20px;
        }

        .form-group label {
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: #333;
        }

        .form-group input {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 1rem;
            transition: border-color 0.3s;
        }

        .form-group input:focus {
            outline: none;
            border-color: var(--primary-start);
        }

        .selected-slot-info {
            background: #e8eaf6;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            text-align: center;
        }

        .selected-slot-info strong {
            color: var(--primary-start);
            font-size: 1.1rem;
        }

        .btn {
            padding: 14px 32px;
            border: none;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s;
            width: 100%;
        }

        .btn-primary {
            background: var(--gradient-bg);
            color: white;
        }

        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }

        .btn-secondary {
            background: #e0e0e0;
            color: #666;
            margin-top: 10px;
        }

        .btn-secondary:hover {
            background: #d0d0d0;
        }

        .loading {
            text-align: center;
            padding: 40px;
            color: #666;
        }

        .month-navigation {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 30px;
            padding: 20px;
            background: #f9f9f9;
            border-radius: 12px;
        }

        .month-title {
            font-size: 1.5rem;
            font-weight: 600;
            color: #333;
        }

        .nav-btn {
            padding: 10px 20px;
            background: var(--primary-start);
            color: white;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-weight: 600;
            transition: all 0.3s;
        }

        .nav-btn:hover {
            background: var(--primary-end);
            transform: translateY(-2px);
        }

        .nav-btn:disabled {
            background: #ccc;
            cursor: not-allowed;
            transform: none;
        }

        .message {
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            text-align: center;
            display: none;
        }

        .message.success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }

        .message.error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }

        .message.active {
            display: block;
        }

        .timezone-info {
            font-size: 0.9rem;
            opacity: 0.85;
            margin-top: 10px;
            padding: 8px 16px;
            background: rgba(255, 255, 255, 0.15);
            border-radius: 6px;
            display: inline-block;
        }

        .duration-toggle {
            display: none; /* Hidden - only 30min slots available */
        }

        .theme-selector {
            display: none; /* Hidden - only Burgundy theme available */
        }

    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Бронювання 30-хвилинної онлайн-консультації з Христиною Івасюк</h1>
            <p>Виберіть день, потім оберіть зручний для вас час</p>
            <div class="timezone-info" id="timezoneInfo">Завантаження часового поясу...</div>
            <div class="duration-toggle">
                <button class="duration-btn active" id="duration30m" onclick="setDuration('30m')">30 хвилин</button>
                <button class="duration-btn" id="duration1h" onclick="setDuration('1h')">1 година</button>
            </div>
            <div class="theme-selector">
                <div class="theme-selector-label">🎨 Виберіть тему:</div>
                <div class="theme-options">
                    <button class="theme-btn active" onclick="changeTheme('rose-garden')" data-theme="rose-garden">
                        <span class="theme-preview rose-garden"></span>Троянда
                    </button>
                    <button class="theme-btn" onclick="changeTheme('cherry-blossom')" data-theme="cherry-blossom">
                        <span class="theme-preview cherry-blossom"></span>Сакура
                    </button>
                    <button class="theme-btn" onclick="changeTheme('sunset-coral')" data-theme="sunset-coral">
                        <span class="theme-preview sunset-coral"></span>Корал
                    </button>
                    <button class="theme-btn" onclick="changeTheme('berry-burst')" data-theme="berry-burst">
                        <span class="theme-preview berry-burst"></span>Ягода
                    </button>
                    <button class="theme-btn" onclick="changeTheme('pink-lemonade')" data-theme="pink-lemonade">
                        <span class="theme-preview pink-lemonade"></span>Рожевий лимонад
                    </button>
                    <button class="theme-btn" onclick="changeTheme('burgundy')" data-theme="burgundy">
                        <span class="theme-preview burgundy"></span>Бургунді
                    </button>
                </div>
            </div>
        </div>

        <div class="content">
            <div id="message" class="message"></div>

            <div id="loading" class="loading">
                Завантаження доступних слотів...
            </div>

            <div id="monthNavigation" class="month-navigation" style="display: none;">
                <button class="nav-btn" id="prevMonth" onclick="changeMonth(-1)">← Попередній</button>
                <div class="month-title" id="currentMonth"></div>
                <button class="nav-btn" id="nextMonth" onclick="changeMonth(1)">Наступний →</button>
            </div>

            <div id="calendar" class="calendar-grid"></div>

            <div id="timeSlotsPanel" class="time-slots-panel">
                <button class="back-btn" onclick="backToCalendar()">← Повернутись до календаря</button>
                <div class="time-slots-header">
                    <h3 id="selectedDateTitle"></h3>
                    <p>Оберіть зручний час</p>
                </div>
                <div id="timeSlotsGrid" class="time-slots-grid"></div>
            </div>

            <div id="bookingForm" class="booking-form">
                <div class="selected-slot-info">
                    <div>Обраний час:</div>
                    <strong id="selectedSlotDisplay"></strong>
                </div>

                <div class="form-group">
                    <label for="name">Ваше ім'я</label>
                    <input type="text" id="name" placeholder="Введіть ваше повне ім'я" required>
                </div>

                <div class="form-group">
                    <label for="email">Ваш email</label>
                    <input type="email" id="email" placeholder="your.email@example.com" required>
                </div>

                <button class="btn btn-primary" onclick="confirmBooking()">Підтвердити бронювання</button>
                <button class="btn btn-secondary" onclick="cancelBooking()">Скасувати</button>
            </div>
        </div>
    </div>

    <script>
        let selectedSlot = null;
        let selectedDay = null;
        let allSlots = [];
        let currentMonthIndex = 0;
        let availableMonths = [];
        let userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

        // Display user's timezone
        function displayTimezone() {
            const tzInfo = document.getElementById('timezoneInfo');
            if (tzInfo) {
                tzInfo.textContent = '🌍 Весь час показано у: ' + userTimezone;
            }
        }

        function formatDateTime(isoString) {
            const date = new Date(isoString);
            const options = {
                weekday: 'long',
                month: 'long',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
                timeZone: userTimezone
            };
            return date.toLocaleString('uk-UA', options);
        }

        function formatDateLong(isoString) {
            const date = new Date(isoString);
            return date.toLocaleDateString('uk-UA', {
                weekday: 'long',
                month: 'long',
                day: 'numeric',
                year: 'numeric',
                timeZone: userTimezone
            });
        }

        function formatTime(isoString) {
            const date = new Date(isoString);
            return date.toLocaleTimeString('uk-UA', {
                hour: '2-digit',
                minute: '2-digit',
                hour12: false,
                timeZone: userTimezone
            });
        }

        function formatMonthYear(year, month) {
            const date = new Date(year, month);
            return date.toLocaleDateString('uk-UA', { month: 'long', year: 'numeric' });
        }

        function getDayKey(dateString) {
            const date = new Date(dateString);
            // Convert to user's timezone for proper day grouping
            const year = parseInt(date.toLocaleString('en-US', { year: 'numeric', timeZone: userTimezone }));
            const month = parseInt(date.toLocaleString('en-US', { month: 'numeric', timeZone: userTimezone }));
            const day = parseInt(date.toLocaleString('en-US', { day: 'numeric', timeZone: userTimezone }));
            return year + '-' + month + '-' + day;
        }

        function groupSlotsByMonth(slots) {
            const monthMap = new Map();

            slots.forEach(slot => {
                const date = new Date(slot.slot_time);
                // Get year and month in user's timezone
                const year = parseInt(date.toLocaleString('en-US', { year: 'numeric', timeZone: userTimezone }));
                const month = parseInt(date.toLocaleString('en-US', { month: 'numeric', timeZone: userTimezone })) - 1;
                const day = parseInt(date.toLocaleString('en-US', { day: 'numeric', timeZone: userTimezone }));
                const key = year + '-' + month;

                if (!monthMap.has(key)) {
                    monthMap.set(key, {
                        year: year,
                        month: month,
                        days: new Map()
                    });
                }

                const dayKey = getDayKey(slot.slot_time);
                const monthData = monthMap.get(key);

                if (!monthData.days.has(dayKey)) {
                    monthData.days.set(dayKey, {
                        date: new Date(year, month, day),
                        slots: []
                    });
                }

                monthData.days.get(dayKey).slots.push(slot);
            });

            return Array.from(monthMap.values()).sort((a, b) => {
                if (a.year !== b.year) return a.year - b.year;
                return a.month - b.month;
            });
        }

        function getMondayBasedWeekday(date) {
            // Convert Sunday (0) to 6, Monday (1) to 0, Tuesday (2) to 1, etc.
            const day = date.getDay();
            return day === 0 ? 6 : day - 1;
        }

        function changeMonth(direction) {
            currentMonthIndex += direction;
            if (currentMonthIndex < 0) currentMonthIndex = 0;
            if (currentMonthIndex >= availableMonths.length) {
                currentMonthIndex = availableMonths.length - 1;
            }
            renderCalendar();
        }

        function updateMonthNavigation() {
            if (availableMonths.length === 0) return;

            const monthData = availableMonths[currentMonthIndex];
            document.getElementById('currentMonth').textContent = formatMonthYear(monthData.year, monthData.month);
            document.getElementById('prevMonth').disabled = currentMonthIndex === 0;
            document.getElementById('nextMonth').disabled = currentMonthIndex === availableMonths.length - 1;
        }

        async function loadSlots() {
            try {
                document.getElementById('loading').style.display = 'block';
                document.getElementById('calendar').style.display = 'none';
                document.getElementById('monthNavigation').style.display = 'none';

                const response = await fetch('/api/slots');
                if (!response.ok) {
                    throw new Error('Failed to load slots');
                }

                allSlots = await response.json();
                availableMonths = groupSlotsByMonth(allSlots);
                currentMonthIndex = 0;

                document.getElementById('loading').style.display = 'none';
                document.getElementById('monthNavigation').style.display = 'flex';
                document.getElementById('calendar').style.display = 'grid';

                renderCalendar();
            } catch (error) {
                console.error('Error loading slots:', error);
                showMessage('Не вдалося завантажити доступні слоти. Будь ласка, оновіть сторінку.', 'error');
                document.getElementById('loading').style.display = 'none';
            }
        }

        function renderCalendar() {
            const calendar = document.getElementById('calendar');
            calendar.innerHTML = '';

            if (availableMonths.length === 0) {
                calendar.innerHTML = '<div style="text-align: center; padding: 40px; color: #666;">Немає доступних слотів.</div>';
                return;
            }

            const monthData = availableMonths[currentMonthIndex];
            const days = Array.from(monthData.days.values()).sort((a, b) => {
                // First sort by date
                const dateDiff = a.date - b.date;
                if (dateDiff !== 0) return dateDiff;
                return 0;
            });

            // Group days by weeks (starting Monday)
            const weeks = [];
            let currentWeek = [];

            days.forEach((dayData, index) => {
                const weekday = getMondayBasedWeekday(dayData.date);

                // If this is Monday (0) and we have days in current week, start a new week
                if (weekday === 0 && currentWeek.length > 0) {
                    weeks.push(currentWeek);
                    currentWeek = [];
                }

                // Add empty cells at the start of the first week if it doesn't start on Monday
                if (index === 0 && weekday > 0) {
                    for (let i = 0; i < weekday; i++) {
                        currentWeek.push(null);
                    }
                }

                currentWeek.push(dayData);
            });

            // Push the last week
            if (currentWeek.length > 0) {
                weeks.push(currentWeek);
            }

            // Render all days from all weeks
            weeks.forEach(week => {
                week.forEach(dayData => {
                    if (dayData === null) {
                        // Empty cell for alignment
                        const emptyDiv = document.createElement('div');
                        emptyDiv.className = 'day-cell no-slots';
                        emptyDiv.style.visibility = 'hidden';
                        calendar.appendChild(emptyDiv);
                        return;
                    }

                    const dayDiv = document.createElement('div');
                    const availableCount = dayData.slots.filter(s => s.available).length;
                    const hasAvailable = availableCount > 0;

                    dayDiv.className = 'day-cell' + (hasAvailable ? '' : ' no-slots');

                    const date = dayData.date;
                    const dayName = date.toLocaleDateString('uk-UA', { weekday: 'short' });
                    const dayNumber = date.getDate();

                    dayDiv.innerHTML = '<div class="day-number">' + dayNumber + '</div>' +
                        '<div class="day-name">' + dayName + '</div>' +
                        '<div class="day-slots-count">' + availableCount + ' доступно</div>';

                    if (hasAvailable) {
                        dayDiv.onclick = () => selectDay(dayData);
                    }

                    calendar.appendChild(dayDiv);
                });
            });

            updateMonthNavigation();
        }

        function selectDay(dayData) {
            selectedDay = dayData;
            document.getElementById('calendar').style.display = 'none';
            document.getElementById('monthNavigation').style.display = 'none';
            document.getElementById('timeSlotsPanel').classList.add('active');

            const dateStr = dayData.date.toLocaleDateString('uk-UA', {
                weekday: 'long',
                month: 'long',
                day: 'numeric',
                year: 'numeric'
            });
            document.getElementById('selectedDateTitle').textContent = dateStr;

            renderTimeSlots(dayData.slots);
            document.getElementById('timeSlotsPanel').scrollIntoView({ behavior: 'smooth' });
        }

        function renderTimeSlots(slots) {
            const grid = document.getElementById('timeSlotsGrid');
            grid.innerHTML = '';

            slots.sort((a, b) => new Date(a.slot_time) - new Date(b.slot_time));

            slots.forEach(slot => {
                const timeSlot = document.createElement('div');
                timeSlot.className = 'time-slot' + (slot.available ? '' : ' booked');

                timeSlot.textContent = formatTime(slot.slot_time);

                if (slot.available) {
                    timeSlot.onclick = () => selectTimeSlot(slot);
                }

                grid.appendChild(timeSlot);
            });
        }

        function selectTimeSlot(slot) {
            selectedSlot = slot;
            document.getElementById('selectedSlotDisplay').textContent = formatDateTime(slot.slot_time);
            document.getElementById('timeSlotsPanel').classList.remove('active');
            document.getElementById('bookingForm').classList.add('active');
            document.getElementById('bookingForm').scrollIntoView({ behavior: 'smooth' });
        }

        function backToCalendar() {
            document.getElementById('timeSlotsPanel').classList.remove('active');
            document.getElementById('calendar').style.display = 'grid';
            document.getElementById('monthNavigation').style.display = 'flex';
            selectedDay = null;
        }

        function cancelBooking() {
            selectedSlot = null;
            document.getElementById('bookingForm').classList.remove('active');
            document.getElementById('name').value = '';
            document.getElementById('email').value = '';

            if (selectedDay) {
                document.getElementById('timeSlotsPanel').classList.add('active');
            } else {
                backToCalendar();
            }
        }

        async function confirmBooking() {
            const name = document.getElementById('name').value.trim();
            const email = document.getElementById('email').value.trim();

            if (!name || !email) {
                showMessage('Будь ласка, заповніть всі поля', 'error');
                return;
            }

            if (!validateEmail(email)) {
                showMessage('Будь ласка, введіть дійсну email адресу', 'error');
                return;
            }

            try {
                const response = await fetch('/api/bookings', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        slot_time: selectedSlot.slot_time,
                        name: name,
                        email: email
                    })
                });

                if (response.ok) {
                    showMessage('Бронювання підтверджено! Ви отримаєте лист-підтвердження найближчим часом.', 'success');

                    // Mark the slot as unavailable in local data
                    markSlotAsBooked(selectedSlot.slot_time);

                    // Clear form and selections
                    const wasSelectedDay = selectedDay;
                    cancelBooking();

                    // Return to calendar view with updated data
                    if (wasSelectedDay) {
                        backToCalendar();
                    }
                    renderCalendar();
                } else if (response.status === 409) {
                    showMessage('Цей слот вже заброньовано. Будь ласка, оберіть інший час.', 'error');
                    loadSlots(); // Reload slots to get fresh data
                } else {
                    const error = await response.text();
                    showMessage('Не вдалося створити бронювання: ' + error, 'error');
                }
            } catch (error) {
                console.error('Error creating booking:', error);
                showMessage('Не вдалося створити бронювання. Будь ласка, спробуйте ще раз.', 'error');
            }
        }

        function markSlotAsBooked(slotTime) {
            // Update in allSlots array
            const slot = allSlots.find(s => s.slot_time === slotTime);
            if (slot) {
                slot.available = false;
            }

            // Update in availableMonths structure
            availableMonths.forEach(monthData => {
                monthData.days.forEach(dayData => {
                    const slotInDay = dayData.slots.find(s => s.slot_time === slotTime);
                    if (slotInDay) {
                        slotInDay.available = false;
                    }
                });
            });
        }

        function validateEmail(email) {
            const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            return re.test(email);
        }

        function showMessage(text, type) {
            const messageDiv = document.getElementById('message');
            messageDiv.textContent = text;
            messageDiv.className = 'message ' + type + ' active';

            setTimeout(() => {
                messageDiv.classList.remove('active');
            }, 5000);
        }

        // Initialize page - display timezone and load slots
        displayTimezone();
        loadSlots();
    </script>
</body>
</html>
`
	fmt.Fprint(w, html)
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="uk">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Панель адміністратора - Календар тренера</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        :root {
            /* Burgundy Theme */
            --primary-start: #800020;
            --primary-end: #5c0011;
            --accent-color: #a0153e;
            --gradient-bg: linear-gradient(135deg, #800020 0%, #5c0011 100%);
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: var(--gradient-bg);
            min-height: 100vh;
            padding: 20px;
            transition: background 0.3s ease;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }

        .header {
            background: var(--gradient-bg);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 2.5rem;
            margin-bottom: 10px;
        }

        .header p {
            font-size: 1.1rem;
            opacity: 0.9;
        }

        .timezone-info {
            font-size: 0.9rem;
            opacity: 0.85;
            margin: 10px 0;
            padding: 8px 16px;
            background: rgba(255, 255, 255, 0.15);
            border-radius: 6px;
            display: inline-block;
        }

        .nav-link {
            display: inline-block;
            margin-top: 15px;
            padding: 10px 20px;
            background: rgba(255,255,255,0.2);
            color: white;
            text-decoration: none;
            border-radius: 6px;
            transition: all 0.3s;
        }

        .nav-link:hover {
            background: rgba(255,255,255,0.3);
        }

        .content {
            padding: 40px;
        }

        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .stat-card {
            background: var(--gradient-bg);
            color: white;
            padding: 25px;
            border-radius: 12px;
            text-align: center;
        }

        .stat-card h3 {
            font-size: 2rem;
            margin-bottom: 10px;
        }

        .stat-card p {
            font-size: 0.9rem;
            opacity: 0.9;
        }

        .filters {
            display: flex;
            gap: 15px;
            margin-bottom: 30px;
            flex-wrap: wrap;
        }

        .filter-btn {
            padding: 12px 24px;
            border: 2px solid #e0e0e0;
            background: white;
            border-radius: 8px;
            cursor: pointer;
            font-weight: 600;
            transition: all 0.3s;
            font-size: 1rem;
        }

        .filter-btn:hover {
            border-color: var(--primary-start);
            background: #f0f4ff;
        }

        .filter-btn.active {
            background: var(--primary-start);
            color: white;
            border-color: var(--primary-start);
        }

        .slots-container {
            background: #f9f9f9;
            border-radius: 12px;
            padding: 30px;
            max-height: 600px;
            overflow-y: auto;
        }

        .slots-grid {
            display: grid;
            gap: 15px;
        }

        .slot-card {
            background: white;
            padding: 20px;
            border-radius: 10px;
            border: 2px solid #e0e0e0;
            display: grid;
            grid-template-columns: 1fr auto;
            align-items: center;
            gap: 20px;
            transition: all 0.3s;
        }

        .slot-card:hover {
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }

        .slot-card.available {
            border-left: 4px solid #4caf50;
        }

        .slot-card.booked {
            border-left: 4px solid #2196f3;
            background: #e3f2fd;
        }

        .slot-card.blocked {
            border-left: 4px solid #f44336;
            background: #ffebee;
        }

        .slot-info h4 {
            font-size: 1.1rem;
            margin-bottom: 8px;
            color: #333;
        }

        .slot-details {
            display: flex;
            gap: 20px;
            flex-wrap: wrap;
            font-size: 0.9rem;
            color: #666;
        }

        .slot-detail {
            display: flex;
            align-items: center;
            gap: 5px;
        }

        .status-badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.8rem;
            font-weight: 600;
            text-transform: uppercase;
        }

        .status-badge.available {
            background: #c8e6c9;
            color: #2e7d32;
        }

        .status-badge.booked {
            background: #bbdefb;
            color: #1565c0;
        }

        .status-badge.blocked {
            background: #ffcdd2;
            color: #c62828;
        }

        .slot-actions {
            display: flex;
            gap: 10px;
        }

        .action-btn {
            padding: 10px 20px;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-weight: 600;
            transition: all 0.3s;
            font-size: 0.9rem;
        }

        .action-btn.block {
            background: #f44336;
            color: white;
        }

        .action-btn.block:hover {
            background: #d32f2f;
        }

        .action-btn.unblock {
            background: #4caf50;
            color: white;
        }

        .action-btn.unblock:hover {
            background: #388e3c;
        }

        .action-btn:disabled {
            background: #ccc;
            cursor: not-allowed;
        }

        .loading {
            text-align: center;
            padding: 40px;
            color: #666;
            font-size: 1.1rem;
        }

        .message {
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            text-align: center;
            display: none;
        }

        .message.success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }

        .message.error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }

        .message.active {
            display: block;
        }

        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #999;
        }

        .empty-state h3 {
            font-size: 1.5rem;
            margin-bottom: 10px;
        }

        /* Modal Styles */
        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            overflow: auto;
            background-color: rgba(0, 0, 0, 0.5);
            animation: fadeIn 0.3s;
        }

        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }

        .modal.active {
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .modal-content {
            background-color: white;
            margin: 20px;
            padding: 0;
            border-radius: 12px;
            width: 90%;
            max-width: 600px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
            animation: slideIn 0.3s;
        }

        @keyframes slideIn {
            from {
                transform: translateY(-50px);
                opacity: 0;
            }
            to {
                transform: translateY(0);
                opacity: 1;
            }
        }

        .modal-header {
            background: var(--gradient-bg);
            color: white;
            padding: 25px 30px;
            border-radius: 12px 12px 0 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .modal-header h2 {
            margin: 0;
            font-size: 1.5rem;
        }

        .close-btn {
            background: transparent;
            border: none;
            color: white;
            font-size: 2rem;
            cursor: pointer;
            padding: 0;
            width: 30px;
            height: 30px;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: 50%;
            transition: background 0.3s;
        }

        .close-btn:hover {
            background: rgba(255, 255, 255, 0.2);
        }

        .modal-body {
            padding: 30px;
        }

        .modal-detail-row {
            margin-bottom: 20px;
            padding-bottom: 20px;
            border-bottom: 1px solid #e0e0e0;
        }

        .modal-detail-row:last-child {
            border-bottom: none;
            margin-bottom: 0;
            padding-bottom: 0;
        }

        .modal-detail-label {
            font-size: 0.85rem;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 8px;
            font-weight: 600;
        }

        .modal-detail-value {
            font-size: 1.1rem;
            color: #333;
            font-weight: 500;
        }

        .modal-detail-value.large {
            font-size: 1.3rem;
            color: var(--primary-start);
            font-weight: 600;
        }

        .slot-card {
            cursor: pointer;
        }

        .slot-card.booked:hover {
            box-shadow: 0 6px 16px rgba(102, 126, 234, 0.2);
            transform: translateY(-2px);
        }

        .theme-selector {
            display: none; /* Hidden - only Burgundy theme available */
        }

    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Панель адміністратора</h1>
            <p>Керування бронюваннями та доступністю слотів</p>
            <div class="timezone-info" id="adminTimezoneInfo">Завантаження часового поясу...</div>
            <a href="/" class="nav-link">← Повернутись до сторінки бронювання</a>
            <div class="theme-selector">
                <div class="theme-selector-label">🎨 Виберіть тему:</div>
                <div class="theme-options">
                    <button class="theme-btn active" onclick="changeTheme('rose-garden')" data-theme="rose-garden">
                        <span class="theme-preview rose-garden"></span>Троянда
                    </button>
                    <button class="theme-btn" onclick="changeTheme('cherry-blossom')" data-theme="cherry-blossom">
                        <span class="theme-preview cherry-blossom"></span>Сакура
                    </button>
                    <button class="theme-btn" onclick="changeTheme('sunset-coral')" data-theme="sunset-coral">
                        <span class="theme-preview sunset-coral"></span>Корал
                    </button>
                    <button class="theme-btn" onclick="changeTheme('berry-burst')" data-theme="berry-burst">
                        <span class="theme-preview berry-burst"></span>Ягода
                    </button>
                    <button class="theme-btn" onclick="changeTheme('pink-lemonade')" data-theme="pink-lemonade">
                        <span class="theme-preview pink-lemonade"></span>Рожевий лимонад
                    </button>
                    <button class="theme-btn" onclick="changeTheme('burgundy')" data-theme="burgundy">
                        <span class="theme-preview burgundy"></span>Бургунді
                    </button>
                </div>
            </div>
        </div>

        <div class="content">
            <div id="message" class="message"></div>

            <div class="stats">
                <div class="stat-card">
                    <h3 id="totalSlots">-</h3>
                    <p>Всього слотів</p>
                </div>
                <div class="stat-card">
                    <h3 id="availableSlots">-</h3>
                    <p>Доступні</p>
                </div>
                <div class="stat-card">
                    <h3 id="bookedSlots">-</h3>
                    <p>Заброньовані</p>
                </div>
                <div class="stat-card">
                    <h3 id="blockedSlots">-</h3>
                    <p>Заблоковані</p>
                </div>
            </div>

            <div class="filters">
                <button class="filter-btn active" onclick="filterSlots('all')">Всі слоти</button>
                <button class="filter-btn" onclick="filterSlots('available')">Доступні</button>
                <button class="filter-btn" onclick="filterSlots('booked')">Заброньовані</button>
                <button class="filter-btn" onclick="filterSlots('blocked')">Заблоковані</button>
            </div>

            <div id="loading" class="loading">
                Завантаження слотів...
            </div>

            <div id="slotsContainer" class="slots-container" style="display: none;">
                <div id="slotsGrid" class="slots-grid"></div>
            </div>
        </div>
    </div>

    <!-- Booking Details Modal -->
    <div id="bookingModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h2>Деталі бронювання</h2>
                <button class="close-btn" onclick="closeModal()">&times;</button>
            </div>
            <div class="modal-body">
                <div class="modal-detail-row">
                    <div class="modal-detail-label">Час зустрічі</div>
                    <div class="modal-detail-value large" id="modalDateTime"></div>
                </div>
                <div class="modal-detail-row">
                    <div class="modal-detail-label">Статус</div>
                    <div class="modal-detail-value">
                        <span class="status-badge booked">ЗАБРОНЬОВАНО</span>
                    </div>
                </div>
                <div class="modal-detail-row">
                    <div class="modal-detail-label">Ім'я клієнта</div>
                    <div class="modal-detail-value" id="modalName"></div>
                </div>
                <div class="modal-detail-row">
                    <div class="modal-detail-label">Email клієнта</div>
                    <div class="modal-detail-value" id="modalEmail"></div>
                </div>
            </div>
        </div>
    </div>

    <script>
        let allSlots = [];
        let currentFilter = 'all';
        let userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

        // Display user's timezone
        function displayTimezone() {
            const tzInfo = document.getElementById('adminTimezoneInfo');
            if (tzInfo) {
                tzInfo.textContent = '🌍 Весь час показано у: ' + userTimezone;
            }
        }

        function formatDateTime(isoString) {
            const date = new Date(isoString);
            return date.toLocaleString('uk-UA', {
                weekday: 'long',
                month: 'long',
                day: 'numeric',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
                timeZone: userTimezone
            });
        }

        function openBookingModal(slot) {
            document.getElementById('modalDateTime').textContent = formatDateTime(slot.slot_time);
            document.getElementById('modalName').textContent = slot.name || 'N/A';
            document.getElementById('modalEmail').textContent = slot.email || 'N/A';
            document.getElementById('bookingModal').classList.add('active');
        }

        function closeModal() {
            document.getElementById('bookingModal').classList.remove('active');
        }

        // Close modal when clicking outside of it
        window.onclick = function(event) {
            const modal = document.getElementById('bookingModal');
            if (event.target === modal) {
                closeModal();
            }
        }

        // Close modal with Escape key
        document.addEventListener('keydown', function(event) {
            if (event.key === 'Escape') {
                closeModal();
            }
        });

        async function loadSlots() {
            try {
                const response = await fetch('/api/admin/slots');
                if (!response.ok) {
                    throw new Error('Failed to load slots');
                }

                allSlots = await response.json();
                updateStats();
                renderSlots();

                document.getElementById('loading').style.display = 'none';
                document.getElementById('slotsContainer').style.display = 'block';
            } catch (error) {
                console.error('Error loading slots:', error);
                showMessage('Не вдалося завантажити слоти. Будь ласка, оновіть сторінку.', 'error');
                document.getElementById('loading').style.display = 'none';
            }
        }

        function updateStats() {
            const stats = {
                total: allSlots.length,
                available: allSlots.filter(s => s.status === 'available').length,
                booked: allSlots.filter(s => s.status === 'booked').length,
                blocked: allSlots.filter(s => s.status === 'blocked').length
            };

            document.getElementById('totalSlots').textContent = stats.total;
            document.getElementById('availableSlots').textContent = stats.available;
            document.getElementById('bookedSlots').textContent = stats.booked;
            document.getElementById('blockedSlots').textContent = stats.blocked;
        }

        function filterSlots(filter) {
            currentFilter = filter;

            // Update button states
            document.querySelectorAll('.filter-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            event.target.classList.add('active');

            renderSlots();
        }

        function renderSlots() {
            const grid = document.getElementById('slotsGrid');
            grid.innerHTML = '';

            const filteredSlots = currentFilter === 'all'
                ? allSlots
                : allSlots.filter(s => s.status === currentFilter);

            if (filteredSlots.length === 0) {
                grid.innerHTML = '<div class="empty-state"><h3>Слотів не знайдено</h3><p>Спробуйте змінити фільтр.</p></div>';
                return;
            }

            filteredSlots.forEach(slot => {
                const slotCard = document.createElement('div');
                slotCard.className = 'slot-card ' + slot.status;

                // Add click handler for booked slots to open modal
                if (slot.status === 'booked') {
                    slotCard.onclick = function() {
                        openBookingModal(slot);
                    };
                }

                let detailsHTML = '<span class="status-badge ' + slot.status + '">' + slot.status + '</span>';

                if (slot.status === 'booked' && slot.name && slot.email) {
                    detailsHTML += '<div class="slot-detail">👤 ' + slot.name + '</div>';
                }

                let actionsHTML = '';
                if (slot.status === 'available') {
                    actionsHTML = '<button class="action-btn block" onclick="event.stopPropagation(); blockSlot(\'' + slot.slot_time + '\')">Заблокувати</button>';
                } else if (slot.status === 'blocked') {
                    actionsHTML = '<button class="action-btn unblock" onclick="event.stopPropagation(); unblockSlot(\'' + slot.slot_time + '\')">Розблокувати</button>';
                }

                slotCard.innerHTML =
                    '<div class="slot-info">' +
                        '<h4>' + formatDateTime(slot.slot_time) + '</h4>' +
                        '<div class="slot-details">' + detailsHTML + '</div>' +
                    '</div>' +
                    '<div class="slot-actions">' + actionsHTML + '</div>';

                grid.appendChild(slotCard);
            });
        }

        async function blockSlot(slotTime) {
            try {
                const response = await fetch('/api/admin/block', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ slot_time: slotTime })
                });

                if (response.ok) {
                    showMessage('Слот успішно заблоковано', 'success');
                    await loadSlots();
                } else {
                    const error = await response.text();
                    showMessage('Не вдалося заблокувати слот: ' + error, 'error');
                }
            } catch (error) {
                console.error('Error blocking slot:', error);
                showMessage('Не вдалося заблокувати слот. Будь ласка, спробуйте ще раз.', 'error');
            }
        }

        async function unblockSlot(slotTime) {
            try {
                const response = await fetch('/api/admin/unblock', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ slot_time: slotTime })
                });

                if (response.ok) {
                    showMessage('Слот успішно розблоковано', 'success');
                    await loadSlots();
                } else {
                    const error = await response.text();
                    showMessage('Не вдалося розблокувати слот: ' + error, 'error');
                }
            } catch (error) {
                console.error('Error unblocking slot:', error);
                showMessage('Не вдалося розблокувати слот. Будь ласка, спробуйте ще раз.', 'error');
            }
        }

        function showMessage(text, type) {
            const messageDiv = document.getElementById('message');
            messageDiv.textContent = text;
            messageDiv.className = 'message ' + type + ' active';

            setTimeout(() => {
                messageDiv.classList.remove('active');
            }, 5000);
        }

        // Theme management
        function changeTheme(themeName) {
            // Update body data-theme attribute
            if (themeName === 'rose-garden') {
                document.body.removeAttribute('data-theme');
            } else {
                document.body.setAttribute('data-theme', themeName);
            }

            // Update active button
            document.querySelectorAll('.theme-btn').forEach(btn => {
                btn.classList.remove('active');
                if (btn.getAttribute('data-theme') === themeName) {
                    btn.classList.add('active');
                }
            });

            // Save preference to localStorage
            localStorage.setItem('selectedTheme', themeName);
        }

        // Load saved theme on page load
        function loadSavedTheme() {
            const savedTheme = localStorage.getItem('selectedTheme') || 'rose-garden';
            changeTheme(savedTheme);
        }

        // Display timezone, load theme and load slots when page loads
        loadSavedTheme();
        displayTimezone();
        loadSlots();
    </script>
</body>
</html>
`
	fmt.Fprint(w, html)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
