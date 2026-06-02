import Ionicons from "@expo/vector-icons/Ionicons";
import { Tabs } from "expo-router/js-tabs";

const ACTIVE = "#3EF090";
const INACTIVE = "#7CA08C";
const TAB_BG = "#0D1C17";
const TAB_BORDER = "#1A2C24";

type IoniconName = keyof typeof Ionicons.glyphMap;

/** Pick the filled glyph when the tab is focused, the outline glyph otherwise. */
function tabIcon(filled: IoniconName, outline: IoniconName) {
  return ({ color, focused, size }: { color: string; focused: boolean; size: number }) => (
    <Ionicons name={focused ? filled : outline} color={color} size={size} />
  );
}

export default function TabLayout() {
  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: ACTIVE,
        tabBarInactiveTintColor: INACTIVE,
        tabBarStyle: { backgroundColor: TAB_BG, borderTopColor: TAB_BORDER },
        tabBarLabelStyle: { fontSize: 11, fontWeight: "500" },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{ title: "Home", tabBarIcon: tabIcon("grid", "grid-outline") }}
      />
      <Tabs.Screen
        name="bible"
        options={{ title: "Bible", tabBarIcon: tabIcon("book", "book-outline") }}
      />
      <Tabs.Screen
        name="stats"
        options={{ title: "Stats", tabBarIcon: tabIcon("stats-chart", "stats-chart-outline") }}
      />
      <Tabs.Screen
        name="community"
        options={{ title: "Community", tabBarIcon: tabIcon("people", "people-outline") }}
      />
      <Tabs.Screen
        name="settings"
        options={{ title: "Settings", tabBarIcon: tabIcon("settings", "settings-outline") }}
      />
    </Tabs>
  );
}
