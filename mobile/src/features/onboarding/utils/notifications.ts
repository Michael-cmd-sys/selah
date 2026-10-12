import * as Notifications from 'expo-notifications';

// Set default notification handler
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: false,
  }),
});

export async function requestNotificationPermissions(): Promise<boolean> {
  try {
    const { status: existingStatus } = await Notifications.getPermissionsAsync();
    let finalStatus = existingStatus;
    if (existingStatus !== 'granted') {
      const { status } = await Notifications.requestPermissionsAsync();
      finalStatus = status;
    }
    return finalStatus === 'granted';
  } catch (error) {
    console.warn('Notification permission error:', error);
    return false;
  }
}

export function parseTimeString(timeStr: string): { hour: number; minute: number } {
  // Format expected: "06:00 AM" or "09:00 PM"
  const match = timeStr.trim().match(/^(\d{1,2}):(\d{2})\s*(AM|PM)$/i);
  if (!match) return { hour: 6, minute: 0 };

  let hour = parseInt(match[1], 10);
  const minute = parseInt(match[2], 10);
  const period = match[3].toUpperCase();

  if (period === 'PM' && hour < 12) hour += 12;
  if (period === 'AM' && hour === 12) hour = 0;

  return { hour, minute };
}

export async function scheduleSuachNotifications(morningTime: string, eveningTime: string) {
  try {
    const granted = await requestNotificationPermissions();
    if (!granted) {
      console.log('Notification permission not granted, skipping notification scheduling.');
      return;
    }

    await Notifications.cancelAllScheduledNotificationsAsync();

    const morning = parseTimeString(morningTime);
    const evening = parseTimeString(eveningTime);

    // Schedule Daily Morning Proverbs
    await Notifications.scheduleNotificationAsync({
      content: {
        title: 'Suach — Morning Proverbs',
        body: 'Time to receive your daily wisdom. Step into peace and abide in Him.',
        sound: true,
      },
      trigger: {
        type: Notifications.SchedulableTriggerInputTypes.DAILY,
        hour: morning.hour,
        minute: morning.minute,
      },
    });

    // Schedule Daily Evening Psalms
    await Notifications.scheduleNotificationAsync({
      content: {
        title: 'Suach — Evening Psalms',
        body: 'Decompress your soul before nightfall. Rest in His presence.',
        sound: true,
      },
      trigger: {
        type: Notifications.SchedulableTriggerInputTypes.DAILY,
        hour: evening.hour,
        minute: evening.minute,
      },
    });

    console.log(`Scheduled daily notifications at Morning: ${morningTime} (${morning.hour}:${morning.minute}) and Evening: ${eveningTime} (${evening.hour}:${evening.minute})`);
  } catch (error) {
    console.warn('Error scheduling notifications:', error);
  }
}
