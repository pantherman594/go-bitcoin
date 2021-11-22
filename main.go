package main

func StartLogger() error {
  return nil
}

func SetupNetwork(numMiners int) error {

  go StartLogger()

  for i := 0; i < numMiners; i++ {
    go StartMiner()
  }

  return nil
}

func main() {
  SetupNetwork(3)
}
