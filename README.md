# comprehensive_exporter

> Nvidai + node exporter for prometheus 

## Nvidia监控项

- Devicecount 显卡数量
- Deviceinfo  显卡驱动版本
- memoryUsed 显存使用
- memoryTotal  显存总量
- powerUsage  功率使用
- temperature  显卡温度
- processInfo  显卡进程
- utilizationGPU 显卡使用率
- utilizationMemory 显存使用率
- utilizationGPUAverage 显卡平均使用率

## 使用

```go
go run main.go
curl http://127.0.0.1:9100/metrics
```

或者编译完执行

```
go build
./comprehensive_exporter
```

