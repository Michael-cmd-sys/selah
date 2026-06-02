import { Redirect, useRouter } from "expo-router";
import { StatusBar } from "expo-status-bar";
import { Pressable, StyleSheet, Text, View } from "react-native";

export default function Onboarding() {
  const router = useRouter();

  // Set this to true to see the onboarding UI,
  // or use logic to determine if onboarding is needed.
  const showOnboarding = false;

  if (!showOnboarding) {
    return <Redirect href="/(tabs)" />;
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>🚀 Breath in/out</Text>

      <Text style={styles.subtitle}>Winning each day with God 🔥</Text>

      <Text style={styles.body}>In whatever you do, don't forget God</Text>

      <Pressable
        style={({ pressed }) => [styles.button, pressed && styles.buttonPressed]}
        onPress={() => router.push("/(tabs)")}
      >
        <Text style={styles.buttonLabel}>Get Started</Text>
      </Pressable>

      <StatusBar style="dark" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "white",
    alignItems: "center",
    justifyContent: "center",
    padding: 32,
    gap: 20,
  },
  title: {
    fontSize: 40,
    fontWeight: "900",
    textAlign: "center",
  },
  subtitle: {
    fontSize: 20,
    textAlign: "center",
    color: "#444",
  },
  body: {
    fontSize: 16,
    textAlign: "center",
    color: "#666",
  },
  button: {
    marginTop: 8,
    paddingVertical: 14,
    paddingHorizontal: 32,
    borderRadius: 14,
    backgroundColor: "#3EF090",
    borderCurve: "continuous",
  },
  buttonPressed: {
    opacity: 0.85,
  },
  buttonLabel: {
    fontSize: 16,
    fontWeight: "700",
    color: "#0D1C17",
  },
});
