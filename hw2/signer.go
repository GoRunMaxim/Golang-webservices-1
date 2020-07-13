package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

var ExecutePipeline = func(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, value := range jobs {
		out := make(chan interface{})
		wg.Add(1)
		go doAllJob(wg, in, out, value)
		in = out
	}
	wg.Wait()
}

func doAllJob(wg *sync.WaitGroup, in, out chan interface{}, jobFunc job) {
	defer wg.Done()
	defer close(out)
	jobFunc(in, out)
}

var SingleHash = func(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for input := range in{
		wg.Add(1)
		data := fmt.Sprintf("%v", input)
		go func() {
			startWorker(data, out, wg, mu)
		}()
	}

	wg.Wait()
}

func startWorker(data string, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex){
	defer wg.Done()
	var wgExtra sync.WaitGroup
	wgExtra.Add(2)
	var a,b,c,sum string
	go func() {
		defer wgExtra.Done()
		mu.Lock()
		a = DataSignerMd5(data)
		mu.Unlock()
		c = DataSignerCrc32(a)
	}()
	go func() {
		defer wgExtra.Done()
		b = DataSignerCrc32(data)
	}()
	wgExtra.Wait()
	sum = b + "~" + c
	out <- sum
}

//crc32(th+data))
var MultiHash = func(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for input := range in{
		wg.Add(1)
		data := fmt.Sprintf("%v", input)
		go func() {
			startMulti(data, out, wg)
		}()
	}
	wg.Wait()
}

func startMulti(data string, out chan interface{}, wg *sync.WaitGroup){
	defer wg.Done()
	var wgExtra sync.WaitGroup
	var b[6] string
	var c string
	wgExtra.Add(6)
	for i := 0; i < 6; i++ {
		go func(i int) {
			defer wgExtra.Done()
			b[i] = DataSignerCrc32(strconv.Itoa(i) + data)
		}(i)
	}
	wgExtra.Wait()
	for i := 0; i < 6; i++ {
		c += b[i]
	}
	out <- c
}

func CombineResults(in, out chan interface{}) {
	var b string
	var trs[] string
	for input := range in{
		trs = append(trs, fmt.Sprintf("%v", input))
	}
	sort.Strings(trs)
	for i, value := range trs{
		if(i != len(trs)-1){
			b+=value + "_"
		}else{
			b+= value
		}
	}
	out <- b
}
