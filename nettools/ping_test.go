// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nettools

// func TestPing_rawPingIcmp4(t *testing.T) {
// 	ctx := context.Background()
// 	target := netip.MustParseAddr("127.0.0.1")
// 	ttl := 2
// 	listenAddress := netip.MustParseAddr("0.0.0.0")
// 	readTimeout := time.Second
// 	icmpID := 1
// 	icmpSeq := 1
// 	allowAllErrors := true
//
// 	resp, err := rawPingIcmp4(
// 		ctx,
// 		target,
// 		ttl,
// 		listenAddress,
// 		readTimeout,
// 		icmpID,
// 		icmpSeq,
// 		allowAllErrors,
// 	)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Printf("respon: %+v\n", resp)
// 	t.Error("break")
// }

// func TestPing_rawPingUdp4(t *testing.T) {
// 	ctx := context.Background()
// 	target := netip.MustParseAddr("127.0.0.1")
// 	ttl := 2
// 	listenAddress := netip.MustParseAddr("0.0.0.0")
// 	readTimeout := time.Second
// 	icmpID := 1
// 	icmpSeq := 1
// 	allowAllErrors := true
//
// 	resp, err := rawPingUdp4(
// 		ctx,
// 		target,
// 		ttl,
// 		listenAddress,
// 		readTimeout,
// 		icmpID,
// 		icmpSeq,
// 		allowAllErrors,
// 	)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if resp.Err != nil {
// 		t.Fatal(err)
// 	}
// }
