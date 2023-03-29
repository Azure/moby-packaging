group default {
  targets = ["buster", "xenial", "bionic", "centos7"]
}

target buster {
  target = "buster"
  platforms = ["linux/amd64", "linux/armhf", "linux/arm64"]
  tags = ["upstreamk8sci.azurecr.io/moby/test-systemd:buster"]
  output = ["type=registry"]
}

target xenial {
  target = "xenial"
  platforms = ["linux/amd64", "linux/armhf", "linux/arm64"]
  tags = ["upstreamk8sci.azurecr.io/moby/test-systemd:xenial"]
  output = ["type=registry"]
}

target bionic {
  target = "bionic"
  platforms = ["linux/amd64", "linux/armhf", "linux/arm64"]
  tags = ["upstreamk8sci.azurecr.io/moby/test-systemd:bionic"]
  output = ["type=registry"]
}

target centos7 {
  target = "centos7"
  platforms = ["linux/amd64", "linux/armhf", "linux/arm64"]
  tags = ["upstreamk8sci.azurecr.io/moby/test-systemd:centos7"]
  output = ["type=registry"]
}

target centos7-test {
  target = "centos7-test"
  platforms = ["linux/amd64"]
}

target mariner2 {
  target = "mariner2"
  platforms = ["linux/amd64"]
  tags = ["upstreamk8sci.azurecr.io/moby/test-systemd:mariner2"]
  output = ["type=registry"]
}

target mariner2-test {
  target = "mariner2-test"
  platforms = ["linux/amd64"]
}

target buster-test {
  target = "buster-test"
  platforms = ["linux/amd64"]
}

target xenial-test {
  target = "xenial-test"
  platforms = ["linux/amd64"]
}

target bionic-test {
  target = "bionic-test"
  platforms = ["linux/amd64"]
}