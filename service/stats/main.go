package stats

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/lavavrik/go-sm/protobufs"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"google.golang.org/protobuf/proto"
)

func GetCompiledStatsPoint() ([]byte, error) {
	// disk io counters
	diskCounters1, err := disk.IOCounters()
	if err != nil {
		return nil, err
	}

	// network io counters
	netCounters1, err := net.IOCounters(false)
	if err != nil {
		return nil, err
	}

	cpuPercents, err := cpu.Percent(1*time.Second, true)
	if err != nil || len(cpuPercents) == 0 {
		return nil, err
	}

	cpuPercentsUint32 := make([]uint32, len(cpuPercents))
	for i, v := range cpuPercents {
		cpuPercentsUint32[i] = uint32(v * 100) // Convert to percentage as uint32
	}

	// memory available and used
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	diskCounters2, err := disk.IOCounters()
	if err != nil {
		return nil, err
	}

	diskCounters := make(map[string]disk.IOCountersStat)
	for name, counter1 := range diskCounters1 {
		if counter2, ok := diskCounters2[name]; ok {
			diskCounters[name] = disk.IOCountersStat{
				ReadBytes:  counter2.ReadBytes - counter1.ReadBytes,
				WriteBytes: counter2.WriteBytes - counter1.WriteBytes,
			}
		} else {
			diskCounters[name] = counter1
		}
	}

	netCounters2, err := net.IOCounters(false)
	if err != nil {
		return nil, err
	}

	netCountersDifferences := net.IOCountersStat{
		BytesSent: netCounters2[0].BytesSent - netCounters1[0].BytesSent,
		BytesRecv: netCounters2[0].BytesRecv - netCounters1[0].BytesRecv,
	}

	data := &protobufs.DataPoint{
		Timestamp: uint64(time.Now().Unix()),
		Cpu: &protobufs.CPU{
			Load: cpuPercentsUint32,
		},
		Memory: &protobufs.Memory{
			Used: memStats.Used,
			Free: memStats.Free,
		},
		Disks: make([]*protobufs.Disk, 0),
		Network: &protobufs.Network{
			BytesSent:     netCountersDifferences.BytesSent,
			BytesReceived: netCountersDifferences.BytesRecv,
		},
	}

	// get disk read and write speeds from diskCounters
	// Speed = bytes transferred during the 1-second measurement interval
	for diskName, counter := range diskCounters {
		var readSpeed, writeSpeed uint64

		// Since we measured over 1 second, the byte differences are already bytes/sec
		readSpeed = counter.ReadBytes
		writeSpeed = counter.WriteBytes

		data.Disks = append(data.Disks, &protobufs.Disk{
			Name:       diskName,
			ReadSpeed:  readSpeed,
			WriteSpeed: writeSpeed,
		})
	}

	return proto.Marshal(data)
}

// Compile multiple DataPoints into a single byte slice
func DataPointsToBytes(points []*protobufs.DataPoint) []byte {
	if len(points) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, p := range points {
		data, err := proto.Marshal(p)
		if err != nil {
			log.Printf("Error marshaling DataPoint: %v", err)
			continue
		}

		// Write the length prefix
		lenBuf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutUvarint(lenBuf, uint64(len(data)))
		buf.Write(lenBuf[:n])
		buf.Write(data)
	}

	return buf.Bytes()
}

func WriteDataPoints() error {
	bytes, err := GetCompiledStatsPoint()
	if err != nil {
		// Handle error
		return err
	}

	// Create a buffer for the varint length
	lenBuf := make([]byte, binary.MaxVarintLen64)
	// Write the length of the message as a varint
	n := binary.PutUvarint(lenBuf, uint64(len(bytes)))

	// Append data to the file in the filesystem using standard GO libraries
	f, err := os.OpenFile("stats.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Handle error
		return err
	}
	defer f.Close()

	// Write the length prefix to the writer
	if _, err := f.Write(lenBuf[:n]); err != nil {
		fmt.Errorf("failed to write length prefix: %w", err)
		return err
	}

	if _, err := f.Write(bytes); err != nil {
		// Handle error
		return err
	}

	return nil
}

func BytesToDataPoints(b []byte) ([]*protobufs.DataPoint, error) {
	reader := bytes.NewReader(b)

	var dataPoints []*protobufs.DataPoint
	for {
		// Create a new DataPoint struct for each message
		dp := &protobufs.DataPoint{}
		err := ReadDelimited(reader, dp)

		if err == io.EOF {
			// We've successfully read all messages
			break
		}
		if err != nil {
			log.Fatalf("Error reading delimited message: %v", err)
		}
		dataPoints = append(dataPoints, dp)
	}

	return dataPoints, nil
}

func ReadDelimited(reader io.Reader, msg proto.Message) error {
	// Read the varint length prefix
	length, err := binary.ReadUvarint(reader.(io.ByteReader))
	if err != nil {
		// io.EOF is an expected error when we're done reading
		return err
	}

	// Read the message bytes
	pBytes := make([]byte, length)
	if _, err := io.ReadFull(reader, pBytes); err != nil {
		return fmt.Errorf("failed to read message bytes: %w", err)
	}

	// Unmarshal the bytes into the provided message struct
	if err := proto.Unmarshal(pBytes, msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}
