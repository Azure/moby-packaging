package main

// // TestDistros makes sure each distro has the appropriate files, no typos or bad copy/pastes when generating the values.
// func TestDistros(t *testing.T) {
// 	for distro, f := range distros {
// 		distro := distro
// 		f := f
// 		t.Run(distro, func(t *testing.T) {
// 			t.Parallel()

// 			c := f(client)
// 			release := c.File("/etc/os-release")

// 			contents, err := release.Contents(context.Background())
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			scanner := bufio.NewScanner(strings.NewReader(contents))
// 			var id string

// 		Scan:
// 			for scanner.Scan() {
// 				k, v, ok := strings.Cut(scanner.Text(), "=")
// 				if !ok {
// 					t.Log("unepxected line with no '='", scanner.Text())
// 					continue
// 				}

// 				v = strings.ReplaceAll(v, `"`, "")
// 				// Depending on the distro there can be different keys.
// 				// All should have ID and VERSION_ID, but PLATFORM_ID is more helpful for rhel based distros.
// 				switch k {
// 				case "ID":
// 					id = v + id
// 				case "VERSION_ID":
// 					id += v
// 				case "PLATFORM_ID":
// 					id = strings.TrimPrefix(v, "platform:")
// 					// This is what we are after with no other values, we can break the loop
// 					break Scan
// 				}
// 			}

// 			if err := scanner.Err(); err != nil {
// 				t.Error(err)
// 			}

// 			if id != distroIDs[distro] {
// 				t.Errorf("expected %q, got %q", distroIDs[distro], id)
// 			}
// 		})
// 	}
// }
