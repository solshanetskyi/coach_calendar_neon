# Cancel Booking Feature

## Overview

Admins can now cancel bookings directly from the admin panel. When a booking is cancelled:
1. The booking is removed from the database
2. The associated Zoom meeting is automatically deleted (if one exists)
3. The slot becomes available again for new bookings

## What Was Added

### 1. Backend - Zoom Meeting Deletion

**File**: [zoom.go:208-319](zoom.go#L208-L319)

Added two new functions:
- `DeleteMeeting(joinURL string) error` - Deletes a Zoom meeting using the Zoom API
- `extractMeetingIDFromURL(joinURL string)` - Helper function to extract meeting ID from Zoom join URL

**How it works**:
- Extracts the meeting ID from the Zoom join URL (e.g., `https://zoom.us/j/1234567890`)
- Authenticates with Zoom API using OAuth token
- Sends DELETE request to Zoom API endpoint
- Handles edge cases (meeting not found, already deleted, etc.)

### 2. Backend - Cancel Booking API Endpoint

**File**: [handlers/api.go:454-525](handlers/api.go#L454-L525)

New endpoint: `POST /api/admin/cancel`

**Request**:
```json
{
  "slot_time": "2026-01-05T10:00:00+01:00"
}
```

**Response** (Success - 200):
```json
{
  "message": "Booking cancelled successfully"
}
```

**Error Responses**:
- `404 Not Found` - Booking doesn't exist
- `500 Internal Server Error` - Database error

**Process**:
1. Parse and validate the slot_time
2. Fetch booking details (including zoom_link) from database
3. Delete Zoom meeting if zoom_link exists
4. Delete booking from database
5. Return success message

**Updated Interface**: [handlers/api.go:19-23](handlers/api.go#L19-L23)
```go
type ZoomMeetingCreator interface {
    CreateMeeting(name, email string, slotTime time.Time) (string, error)
    DeleteMeeting(joinURL string) error  // NEW
}
```

### 3. Frontend - Admin UI Updates

**File**: [handlers/pages.go](handlers/pages.go)

#### Added Cancel Button
- **Line 1544-1545**: Added cancel button to booked slots in the slot grid
- **Button color**: Orange (`#ff9800`) to distinguish from block (red) and unblock (green)

#### Added JavaScript Function
- **Line 1605-1630**: `cancelBooking(slotTime, customerName)` function
- Shows confirmation dialog before cancelling
- Displays customer name in confirmation
- Warns that Zoom meeting will also be deleted
- Calls `/api/admin/cancel` API endpoint
- Shows success/error messages

#### Added CSS Styling
- **Line 1129-1136**: Styles for `.action-btn.cancel`
- Orange background with darker orange on hover
- Consistent with existing action button styles

### 4. Route Registration

**File**: [main.go:162](main.go#L162)

Added route:
```go
http.HandleFunc("/api/admin/cancel", apiHandlers.CancelBooking)
```

## Usage

### Admin Panel

1. Navigate to `/admin`
2. Find a booked slot (shows in the grid with customer name)
3. Click the **"Скасувати"** (Cancel) button on the booking card
4. Confirm the cancellation in the dialog
5. The booking is cancelled and:
   - Slot becomes available again
   - Zoom meeting is deleted (if it existed)
   - Success message is displayed

### API Usage

You can also cancel bookings programmatically:

```bash
curl -X POST http://localhost:8080/api/admin/cancel \
  -H "Content-Type: application/json" \
  -d '{"slot_time": "2026-01-05T10:00:00+01:00"}'
```

## Features

### ✅ Confirmation Dialog
Before cancelling, admins see:
```
Ви впевнені, що хочете скасувати бронювання для [Customer Name]?

Це також видалить Zoom зустріч, якщо вона існує.
```

### ✅ Automatic Zoom Deletion
- If the booking has a Zoom meeting, it's automatically deleted
- Zoom deletion errors are logged but don't fail the cancellation
- If Zoom integration is disabled, cancellation still works

### ✅ Error Handling
- Invalid slot time format → 400 Bad Request
- Booking not found → 404 Not Found
- Database errors → 500 Internal Server Error
- Zoom API errors → Logged but booking still cancelled

### ✅ Visual Feedback
- Success message: "Бронювання успішно скасовано" (green)
- Error message: "Не вдалося скасувати бронювання: [error]" (red)
- Slot grid automatically refreshes after cancellation

## Security Considerations

1. **Admin-only endpoint**: This endpoint should be protected by authentication (to be added)
2. **Confirmation required**: UI requires confirmation before cancelling
3. **Audit logging**: Cancellations are logged with customer details
4. **No email notification**: Currently doesn't notify customer (future enhancement)

## Future Enhancements

Potential improvements:
- [ ] Add authentication/authorization for admin endpoints
- [ ] Send cancellation email to customer
- [ ] Add reason field for cancellation
- [ ] Store cancellation history (soft delete instead of hard delete)
- [ ] Undo cancellation feature
- [ ] Bulk cancellation support

## Testing

The feature has been tested:
- ✅ Application builds successfully
- ✅ Server starts without errors
- ✅ API endpoint is registered
- ✅ Zoom deletion function compiles
- ✅ UI includes cancel button and styling

To test the full flow:
1. Create a test booking with Zoom meeting
2. Cancel the booking from admin panel
3. Verify booking is removed from database
4. Verify Zoom meeting is deleted (if Zoom is enabled)

## Files Modified

1. **[zoom.go](zoom.go)** - Added DeleteMeeting functionality
2. **[handlers/api.go](handlers/api.go)** - Added CancelBooking endpoint
3. **[handlers/pages.go](handlers/pages.go)** - Added cancel button and UI
4. **[main.go](main.go)** - Registered cancel endpoint

## Summary

The cancel booking feature is complete and ready to use. Admins can now:
- Cancel bookings with a single click
- Automatically delete associated Zoom meetings
- See clear confirmation and success messages
- Have bookings immediately returned to available status
