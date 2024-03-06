package executor

import (
	"errors"
	"sync"
)

type Task func()

type ThreadPool struct {
	workerCroe int
	workerMax  int
	taskQueue  chan Task
	wg         *sync.WaitGroup
	quit       chan bool
	workers    int
}

const Core = "Core"
const Ordinary = "Ordinary"

func NewThreadPool(workerCore, workerMax, maxTaskNum int) (*ThreadPool, error) {
	if workerMax < workerCore {
		return nil, errors.New("workerMax must be greater than or equal to workerCore")
	}
	pool := &ThreadPool{
		workerCroe: workerCore,
		workerMax:  workerMax,
		taskQueue:  make(chan Task, maxTaskNum),
		wg:         &sync.WaitGroup{},
	}

	for i := 0; i < workerCore; i++ {
		go pool.worker(Core)
		pool.workers++
	}

	return pool, nil
}

func (p *ThreadPool) worker(Type string) {
	for {
		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				return // 任务队列关闭，安全退出
			}
			task()      // 执行任务
			p.wg.Done() // 任务完成，减少等待组计数
			if Type == Ordinary && len(p.taskQueue) < cap(p.taskQueue)/2 {
				return // 任务队列小于一半时，普通任务安全退出
			}
		default:
			if Type == Ordinary {
				return // 任务队列为空，普通任务安全退出
			}
			task, ok := <-p.taskQueue
			if !ok {
				return // 任务队列关闭，安全退出
			}
			task()      // 执行任务
			p.wg.Done() // 任务完成，减少等待组计数
		}
	}
}

func (p *ThreadPool) AddTask(task Task) (err error) {
	p.wg.Add(1)
	select {
	case p.taskQueue <- task: // 将任务添加到队列
		return err
	default:
		// 任务队列装满后,开启普通任务
		if p.workers < p.workerMax {
			p.workers++
			go p.worker(Ordinary)
		}
		p.taskQueue <- task // 将任务添加到队列
		return err
	}

}

func (p *ThreadPool) Wait() {
	p.wg.Wait() // 等待所有任务完成
}
