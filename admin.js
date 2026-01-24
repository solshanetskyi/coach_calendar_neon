// Global variables
let allSlots = [];
let allClients = [];
let currentFilter = 'all';
let currentClientId = null;
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

// Tab switching
function switchTab(tabName) {
    // Update tab buttons
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.remove('active');
    });
    event.target.classList.add('active');

    // Update tab content
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById(tabName + '-tab').classList.add('active');

    // Load data for the tab
    if (tabName === 'calendar') {
        loadSlots();
    } else if (tabName === 'clients') {
        loadClients();
    }
}

// ===== CALENDAR TAB FUNCTIONS =====

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
        } else if (slot.status === 'booked') {
            // Escape single quotes in name, email, and phone for use in onclick
            const escapedName = (slot.name || '').replace(/'/g, "\\'");
            const escapedEmail = (slot.email || '').replace(/'/g, "\\'");
            const escapedPhone = (slot.phone || '').replace(/'/g, "\\'");
            actionsHTML = '<button class="action-btn cancel" onclick="event.stopPropagation(); cancelBooking(\'' + slot.slot_time + '\', \'' + escapedName + '\')">Скасувати</button>';
            actionsHTML += '<button class="action-btn edit" onclick="event.stopPropagation(); createClientFromBooking(\'' + escapedName + '\', \'' + escapedEmail + '\', \'' + escapedPhone + '\')">Створити клієнта</button>';
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

function openBookingModal(slot) {
    document.getElementById('modalDateTime').textContent = formatDateTime(slot.slot_time);
    document.getElementById('modalName').textContent = slot.name || 'N/A';
    document.getElementById('modalEmail').textContent = slot.email || 'N/A';
    document.getElementById('modalPhone').textContent = slot.phone || 'Не вказано';
    document.getElementById('bookingModal').classList.add('active');
}

function closeBookingModal() {
    document.getElementById('bookingModal').classList.remove('active');
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

async function cancelBooking(slotTime, customerName) {
    if (!confirm('Ви впевнені, що хочете скасувати бронювання для ' + customerName + '?\n\nЦе також видалить Zoom зустріч, якщо вона існує.')) {
        return;
    }

    try {
        const response = await fetch('/api/admin/cancel', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ slot_time: slotTime })
        });

        if (response.ok) {
            showMessage('Бронювання успішно скасовано', 'success');
            await loadSlots();
        } else {
            const error = await response.text();
            showMessage('Не вдалося скасувати бронювання: ' + error, 'error');
        }
    } catch (error) {
        console.error('Error cancelling booking:', error);
        showMessage('Не вдалося скасувати бронювання. Будь ласка, спробуйте ще раз.', 'error');
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

// ===== CLIENTS TAB FUNCTIONS =====

async function loadClients() {
    try {
        document.getElementById('clientsLoading').style.display = 'block';
        document.getElementById('clientsContainer').style.display = 'none';

        const response = await fetch('/api/admin/clients');
        if (!response.ok) {
            throw new Error('Failed to load clients');
        }

        allClients = await response.json();
        renderClients();

        document.getElementById('clientsLoading').style.display = 'none';
        document.getElementById('clientsContainer').style.display = 'grid';
    } catch (error) {
        console.error('Error loading clients:', error);
        showClientsMessage('Не вдалося завантажити клієнтів. Будь ласка, оновіть сторінку.', 'error');
        document.getElementById('clientsLoading').style.display = 'none';
    }
}

function renderClients() {
    const container = document.getElementById('clientsContainer');
    container.innerHTML = '';

    if (allClients.length === 0) {
        container.innerHTML = '<div class="empty-state"><h3>Клієнтів не знайдено</h3><p>Додайте першого клієнта, натиснувши кнопку вище.</p></div>';
        return;
    }

    allClients.forEach(client => {
        const clientCard = document.createElement('div');
        clientCard.className = 'client-card';

        let infoHTML = '';
        if (client.email) {
            infoHTML += '<div class="client-info-row">📧 ' + client.email + '</div>';
        }
        if (client.phone_number) {
            infoHTML += '<div class="client-info-row">📞 ' + client.phone_number + '</div>';
        }
        if (client.telegram_id) {
            infoHTML += '<div class="client-info-row">💬 ' + client.telegram_id + '</div>';
        }

        let notesHTML = '';
        if (client.notes) {
            notesHTML = '<div class="client-notes"><strong>Нотатки:</strong><br>' + client.notes + '</div>';
        }

        clientCard.innerHTML = `
            <div class="client-header">
                <div class="client-name">${client.full_name}</div>
                <div class="client-actions">
                    <button class="action-btn edit" onclick="editClient(${client.id})">Редагувати</button>
                    <button class="action-btn delete" onclick="deleteClient(${client.id}, '${client.full_name}')">Видалити</button>
                </div>
            </div>
            <div class="client-info">
                ${infoHTML}
            </div>
            ${notesHTML}
            <button class="action-btn create-appointment" onclick="createAppointmentForClient(${client.id}, '${client.full_name}')">
                📅 Створити запис на консультацію
            </button>
        `;

        container.appendChild(clientCard);
    });
}

function openClientModal(clientId = null, prefilledData = null) {
    currentClientId = clientId;
    const modal = document.getElementById('clientModal');
    const form = document.getElementById('clientForm');

    if (clientId) {
        // Edit mode
        const client = allClients.find(c => c.id === clientId);
        if (!client) return;

        document.getElementById('clientModalTitle').textContent = 'Редагувати клієнта';
        document.getElementById('clientId').value = client.id;
        document.getElementById('clientFullName').value = client.full_name;
        document.getElementById('clientEmail').value = client.email || '';
        document.getElementById('clientPhone').value = client.phone_number || '';
        document.getElementById('clientTelegram').value = client.telegram_id || '';
        document.getElementById('clientNotes').value = client.notes || '';
    } else {
        // Create mode
        document.getElementById('clientModalTitle').textContent = 'Додати клієнта';
        form.reset();

        // Pre-fill with provided data if available
        if (prefilledData) {
            if (prefilledData.full_name) {
                document.getElementById('clientFullName').value = prefilledData.full_name;
            }
            if (prefilledData.email) {
                document.getElementById('clientEmail').value = prefilledData.email;
            }
            if (prefilledData.phone_number) {
                document.getElementById('clientPhone').value = prefilledData.phone_number;
            }
            if (prefilledData.telegram_id) {
                document.getElementById('clientTelegram').value = prefilledData.telegram_id;
            }
            if (prefilledData.notes) {
                document.getElementById('clientNotes').value = prefilledData.notes;
            }
        }
    }

    modal.classList.add('active');
}

function closeClientModal() {
    document.getElementById('clientModal').classList.remove('active');
    document.getElementById('clientForm').reset();
    currentClientId = null;
}

function editClient(clientId) {
    openClientModal(clientId);
}

function createClientFromBooking(name, email, phone) {
    // Open the client modal with pre-filled data from the booking
    openClientModal(null, {
        full_name: name,
        email: email,
        phone_number: phone
    });

    // Show a message to indicate the form is pre-filled
    showClientsMessage('Форма клієнта заповнена даними з бронювання', 'success');
}

async function deleteClient(clientId, clientName) {
    if (!confirm('Ви впевнені, що хочете видалити клієнта "' + clientName + '"?')) {
        return;
    }

    try {
        const response = await fetch('/api/admin/clients?id=' + clientId, {
            method: 'DELETE'
        });

        if (response.ok) {
            showClientsMessage('Клієнта успішно видалено', 'success');
            await loadClients();
        } else {
            const error = await response.text();
            showClientsMessage('Не вдалося видалити клієнта: ' + error, 'error');
        }
    } catch (error) {
        console.error('Error deleting client:', error);
        showClientsMessage('Не вдалося видалити клієнта. Будь ласка, спробуйте ще раз.', 'error');
    }
}

function createAppointmentForClient(clientId, clientName) {
    // For now, just show an alert. In the future, this could open a booking modal
    // pre-filled with the client's information
    alert('Функція створення запису для "' + clientName + '" буде реалізована найближчим часом.\n\nПоки що, будь ласка, використовуйте головну сторінку бронювання.');
}

function showClientsMessage(text, type) {
    const messageDiv = document.getElementById('clientsMessage');
    messageDiv.textContent = text;
    messageDiv.className = 'message ' + type + ' active';

    setTimeout(() => {
        messageDiv.classList.remove('active');
    }, 5000);
}

// Client form submission
document.getElementById('clientForm').addEventListener('submit', async function(e) {
    e.preventDefault();

    const clientData = {
        full_name: document.getElementById('clientFullName').value,
        email: document.getElementById('clientEmail').value,
        phone_number: document.getElementById('clientPhone').value,
        telegram_id: document.getElementById('clientTelegram').value,
        notes: document.getElementById('clientNotes').value
    };

    try {
        let response;
        if (currentClientId) {
            // Update existing client
            response = await fetch('/api/admin/clients?id=' + currentClientId, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(clientData)
            });
        } else {
            // Create new client
            response = await fetch('/api/admin/clients', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(clientData)
            });
        }

        if (response.ok) {
            showClientsMessage(
                currentClientId ? 'Клієнта успішно оновлено' : 'Клієнта успішно додано',
                'success'
            );
            closeClientModal();
            await loadClients();
        } else {
            const error = await response.text();
            showClientsMessage('Помилка: ' + error, 'error');
        }
    } catch (error) {
        console.error('Error saving client:', error);
        showClientsMessage('Не вдалося зберегти клієнта. Будь ласка, спробуйте ще раз.', 'error');
    }
});

// Close modals when clicking outside
window.onclick = function(event) {
    const bookingModal = document.getElementById('bookingModal');
    const clientModal = document.getElementById('clientModal');

    if (event.target === bookingModal) {
        closeBookingModal();
    }
    if (event.target === clientModal) {
        closeClientModal();
    }
}

// Close modals with Escape key
document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
        closeBookingModal();
        closeClientModal();
    }
});

// Initialize page
displayTimezone();
loadSlots();
