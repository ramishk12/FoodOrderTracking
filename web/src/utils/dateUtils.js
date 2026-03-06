/**
 * Format a UTC date to PST (Pacific Standard Time) timezone
 * @param {string|Date} dateString - ISO date string or Date object
 * @param {string} format - 'full' (date + time) or 'date' (date only)
 * @returns {string} Formatted date in PST timezone
 */
export const formatDateInPST = (dateString, format = 'full') => {
  if (!dateString) return '';
  
  try {
    const date = new Date(dateString);
    
    if (format === 'full') {
      // Display full date and time in PST
      return date.toLocaleString('en-US', {
        timeZone: 'America/Los_Angeles',
        year: 'numeric',
        month: 'numeric',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: true
      });
    } else if (format === 'date') {
      // Display date only in PST
      return date.toLocaleDateString('en-US', {
        timeZone: 'America/Los_Angeles',
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      });
    } else if (format === 'datetime-local') {
      // Convert to datetime-local format for input fields (YYYY-MM-DDTHH:mm)
      const pstDate = new Date(date.toLocaleString('en-US', { timeZone: 'America/Los_Angeles' }));
      const year = pstDate.getFullYear();
      const month = String(pstDate.getMonth() + 1).padStart(2, '0');
      const day = String(pstDate.getDate()).padStart(2, '0');
      const hours = String(pstDate.getHours()).padStart(2, '0');
      const minutes = String(pstDate.getMinutes()).padStart(2, '0');
      return `${year}-${month}-${day}T${hours}:${minutes}`;
    }
  } catch (error) {
    console.error('Error formatting date:', error);
    return '';
  }
};

/**
 * Get the date status for scheduling (Today, Tomorrow, This Week, Overdue)
 * @param {string|Date} dateString - ISO date string or Date object
 * @returns {string} Status string
 */
export const getOrderDateStatus = (dateString) => {
  if (!dateString) return '';
  
  try {
    const scheduledDate = new Date(dateString);
    const today = new Date();
    
    // Normalize times to compare dates only (in PST)
    const scheduledDatePST = new Date(scheduledDate.toLocaleString('en-US', { timeZone: 'America/Los_Angeles' }));
    const todayPST = new Date(today.toLocaleString('en-US', { timeZone: 'America/Los_Angeles' }));
    
    scheduledDatePST.setHours(0, 0, 0, 0);
    todayPST.setHours(0, 0, 0, 0);
    
    const timeDiff = scheduledDatePST - todayPST;
    const daysDiff = Math.ceil(timeDiff / (1000 * 3600 * 24));
    
    if (daysDiff < 0) {
      return 'overdue';
    } else if (daysDiff === 0) {
      return 'today';
    } else if (daysDiff === 1) {
      return 'tomorrow';
    } else if (daysDiff <= 7) {
      return 'this-week';
    } else {
      return 'future';
    }
  } catch (error) {
    console.error('Error calculating date status:', error);
    return '';
  }
};
