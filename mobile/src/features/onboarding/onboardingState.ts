import { observable } from '@legendapp/state';
import AsyncStorage from '@react-native-async-storage/async-storage';

export interface OnboardingState {
  currentStep: number;             // 1: Silence Test, 2: Diagnosis, 3: Class & Time Config, 4: Narrative, 5: Covenant Lock
  isHolding: boolean;              // Tracking active press state for Step 1
  secondsRemaining: number;          // Silence countdown timer
  userDiagnosis: string;             // Stores user response string from Step 2
  morningTime: string;               // Morning Proverbs target time
  eveningTime: string;               // Evening Psalms target time
  consistencyTarget: number;         // Rolling monthly target percentage (50%, 70%, 85%, 95%, 100%)
  tribalClass: string;               // Dynamic string mapping for the user's tribe identity
  isCommitted: boolean;              // Complete flag to trigger navigation reset to App Home
}

export interface TribalInfo {
  classGroup: string;
  scripture: string;
  vision: string;
  targetDays: string;
}

export const TRIBAL_CLASSES: Record<number, TribalInfo> = {
  50: {
    classGroup: 'CLASS: THE SEEKER',
    scripture: 'Jeremiah 29:13 — "You will seek me and find me when you seek me with all your heart."',
    vision: 'Targets 15 out of 30 days of active fellowship. This class strips away all systemic pressure. Perfect for absolute beginners looking to taste, explore, and establish a brand-new habit footprint. Welcome to the Seekers.',
    targetDays: '15 / 30 Days',
  },
  70: {
    classGroup: 'CLASS: THE BRANCH',
    scripture: 'John 15:5 — "For my God is my vine and I am the branches..."',
    vision: "Targets roughly 21 out of 30 days of active fellowship. Perfect for users moving past the exploration phase who want to build a stable foundation. Grace is heavily integrated to absorb life's sudden disruptions.",
    targetDays: '21 / 30 Days',
  },
  85: {
    classGroup: 'CLASS: THE FELLOWSHIP',
    scripture: 'Amos 3:3 — "Do two walk together unless they have agreed to do so?"',
    vision: 'Targets roughly 25 out of 30 days. High consistency. For individuals ready to build momentum and walk in absolute harmony and daily agreement with their morning and evening focus windows.',
    targetDays: '25 / 30 Days',
  },
  95: {
    classGroup: 'CLASS: THE SCHOLAR',
    scripture: '2 Tim 2:15 — "Study to show thyself approved unto God..."',
    vision: 'Targets roughly 28 out of 30 days. Rigorous calibration. For deep-sea divers who treat daily meditation as an essential priority alongside intense academic and professional workloads.',
    targetDays: '28 / 30 Days',
  },
  100: {
    classGroup: 'CLASS: THE DANIEL COHORT',
    scripture: 'Daniel 6:10 — "An excellent spirit was found in him... he prayed three times a day."',
    vision: "Targets a full 30 out of 30 days. Unwavering, unshakeable execution. Standing firm against the frantic pace of the world. Suach's Grace System will freeze your count if a link drops, but your target remains the full harvest.",
    targetDays: '30 / 30 Days',
  },
};

export const onboardingState$ = observable<OnboardingState>({
  currentStep: 1,
  isHolding: false,
  secondsRemaining: 10,
  userDiagnosis: '',
  morningTime: '06:00 AM',
  eveningTime: '09:00 PM',
  consistencyTarget: 70,
  tribalClass: 'THE BRANCH',
  isCommitted: false,
});

const PERSISTENCE_KEY = 'suach_onboarding_config_v1';

export async function saveOnboardingStateToStorage() {
  try {
    const data = {
      userDiagnosis: onboardingState$.userDiagnosis.get(),
      morningTime: onboardingState$.morningTime.get(),
      eveningTime: onboardingState$.eveningTime.get(),
      consistencyTarget: onboardingState$.consistencyTarget.get(),
      tribalClass: onboardingState$.tribalClass.get(),
      isCommitted: true,
      completedAt: new Date().toISOString(),
    };
    await AsyncStorage.setItem(PERSISTENCE_KEY, JSON.stringify(data));
    onboardingState$.isCommitted.set(true);
  } catch (e) {
    console.error('Error saving onboarding data to storage:', e);
  }
}

export async function loadSavedOnboardingState(): Promise<boolean> {
  try {
    const saved = await AsyncStorage.getItem(PERSISTENCE_KEY);
    if (saved) {
      const data = JSON.parse(saved);
      if (data && data.isCommitted) {
        onboardingState$.userDiagnosis.set(data.userDiagnosis || '');
        onboardingState$.morningTime.set(data.morningTime || '06:00 AM');
        onboardingState$.eveningTime.set(data.eveningTime || '09:00 PM');
        onboardingState$.consistencyTarget.set(data.consistencyTarget || 70);
        onboardingState$.tribalClass.set(data.tribalClass || 'THE BRANCH');
        onboardingState$.isCommitted.set(true);
        return true;
      }
    }
  } catch (e) {
    console.error('Error loading onboarding state:', e);
  }
  return false;
}
