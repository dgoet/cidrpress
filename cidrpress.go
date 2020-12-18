package main

import (
    "net"
    "flag"
    "log"
    "os"
    "bufio"
    "fmt"
    "strings"
    "math"
)

func writeBucketToFile(curBucket []net.IP, filename string, bucketNum int) {
    fullFilename := fmt.Sprintf("%v%v",filename,bucketNum)
    outFile, err := os.Create(fullFilename)
    if err != nil {
        log.Fatalf("Unable to create file: %v\n", fullFilename)
    }
    defer outFile.Close()

    for _, ip := range(curBucket) {
        fmt.Fprintln(outFile, ip.String())
    }
}

func expandCIDR(cidr net.IPNet) []net.IP {
    ones, bits := cidr.Mask.Size()
    totalIPs := int(math.Pow(2,float64(bits-ones)))
    cidrIPs := make([]net.IP, totalIPs)
    curIP := cidr.IP.Mask(cidr.Mask)
    for i:=0; i<totalIPs; i++ {
        copyIP := make(net.IP, len(curIP))
        copy(copyIP, curIP)
        cidrIPs[i] = copyIP
        
        // increment the ip
        bCur := len(curIP)-1
        curIP[bCur]++
        for curIP[bCur] == 0 && bCur > 0 {
            bCur--
            curIP[bCur]++
        }
    }
    return cidrIPs
}

func main() {
    ifPtr := flag.String("if", "all_ips", "File name for list of input IPs")
    bsPtr := flag.Int("bs", 1024, "Bucket size, max number of IPs per file")
    ofPtr := flag.String("of", "ip_bucket_", "Output filename, will be suffixed with the bucket number")
    flag.Parse()
    
    if *bsPtr <= 0 {
        log.Fatal("Bucket size must be greater than 0")
    }

    curBucketNum := 0
    curBucket := make([]net.IP, 0)
    inFile,err := os.Open(*ifPtr)
    if err != nil {
        log.Fatalf("Failed to open input file: 5v\n", *ifPtr)
    }
    defer inFile.Close()
    scanner := bufio.NewScanner(inFile)
    scanner.Split(bufio.ScanLines)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.Contains(line, "/") {
            _, ipNet, err := net.ParseCIDR(line)
            if err != nil {
                log.Fatalf("failed to parse line: %v\n", line)
            }
            
            cidrIPs := expandCIDR(*ipNet)
            for _, ip := range cidrIPs {
                curBucket = append(curBucket, ip)
                if len(curBucket) >= *bsPtr {
                    writeBucketToFile(curBucket, *ofPtr, curBucketNum)
                    curBucket = curBucket[:0]
                    curBucketNum++
                }
            }
        } else {
            ip := net.ParseIP(line)
            if ip == nil {
                log.Fatalf("failed to parse line: %v\n", line)
            }
            curBucket = append(curBucket, ip)
            if len(curBucket) >= *bsPtr {
                writeBucketToFile(curBucket, *ofPtr, curBucketNum)
                curBucket = curBucket[:0]
                curBucketNum++
            }
        }

    }
    if len(curBucket) != 0 {
        writeBucketToFile(curBucket, *ofPtr, curBucketNum)
    }
}
