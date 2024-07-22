1、安装ffmpeg  
2、导入包  
`
import "github.com/H1d3rOne/BcutASR"  \
`
3、使用方法  
`  
client := BcutASR.NewClient()  
client.Input("audio.mp3")  
client.Format("srt")  
client.Output("subtitle.srt")  
client.Run()  
`


