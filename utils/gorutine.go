package utils

import (
	"context"

	clog "docker-server-mgr/utils/log" //custom log
	"time"
)

/*
function를 고루틴에서 실행하고,
panic 시 복구한 뒤 일정 대기 후 재실행을 무한 반복.
ctx가 취소되면 루프를 종료함.
*/
func SafeGoRoutineCtx(ctx context.Context, f func()) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				// 컨텍스트 취소 시 루프 종료
				return
			default:
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						clog.Error("Recovered from panic: %v", r)
					}
				}()
				// 실제 작업 함수
				f()
			}()

			clog.Error("SafeGoRoutine: function exited, restarting in 1s...")
			// 재시작 전 잠시 대기
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}()
}
